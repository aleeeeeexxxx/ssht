package tunnel

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/aleeeeeexxxx/ssht/internal/logger"
	"golang.org/x/crypto/ssh"
)

type State int

const (
	StateStopped State = iota
	StateConnecting
	StateConnected
	StateReconnecting
	StateError
)

const maxRetries = 5
const keepaliveInterval = 30 * time.Second

func (s State) String() string {
	switch s {
	case StateStopped:
		return "Stopped"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateReconnecting:
		return "Reconnecting"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

type Config struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	AuthMethod string `json:"auth_method"` // "key" or "password"
	KeyPath    string `json:"key_path"`
	Password   string `json:"password,omitempty"`
	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`
	LocalHost  string `json:"local_host"`
	LocalPort  int    `json:"local_port"`
}

func (c *Config) SSHAddr() string {
	port := c.Port
	if port == 0 {
		port = 22
	}
	return fmt.Sprintf("%s:%d", c.Host, port)
}

func (c *Config) RemoteAddr() string {
	return fmt.Sprintf("%s:%d", c.RemoteHost, c.RemotePort)
}

func (c *Config) LocalAddr() string {
	host := c.LocalHost
	if host == "" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("%s:%d", host, c.LocalPort)
}

type StateChange struct {
	Name     string
	State    State
	ErrorMsg string
}

type Tunnel struct {
	Config            Config
	state             State
	errorMsg          string
	mu                sync.RWMutex
	client            *ssh.Client
	listener          net.Listener
	ctx               context.Context
	cancel            context.CancelFunc
	keepaliveCancel   context.CancelFunc
	stateChanges      chan<- StateChange
	reconnectInterval time.Duration
}

func New(cfg Config, stateChanges chan<- StateChange) *Tunnel {
	return &Tunnel{
		Config:            cfg,
		state:             StateStopped,
		stateChanges:      stateChanges,
		reconnectInterval: 5 * time.Second,
	}
}

func (t *Tunnel) State() State {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.state
}

func (t *Tunnel) ErrorMsg() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.errorMsg
}

func (t *Tunnel) setState(s State, errMsg string) {
	t.mu.Lock()
	t.state = s
	t.errorMsg = errMsg
	t.mu.Unlock()

	logger.Log.Infow("tunnel state changed",
		"tunnel", t.Config.Name,
		"state", s.String(),
		"error", errMsg,
	)

	if t.stateChanges != nil {
		t.stateChanges <- StateChange{
			Name:     t.Config.Name,
			State:    s,
			ErrorMsg: errMsg,
		}
	}
}

func (t *Tunnel) Start() error {
	t.mu.Lock()
	if t.state == StateConnected || t.state == StateConnecting || t.state == StateReconnecting {
		t.mu.Unlock()
		return fmt.Errorf("tunnel already running")
	}
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.mu.Unlock()

	logger.Log.Infow("starting tunnel",
		"tunnel", t.Config.Name,
		"ssh", t.Config.SSHAddr(),
		"remote", t.Config.RemoteAddr(),
		"local", t.Config.LocalAddr(),
	)

	go t.run()
	return nil
}

func (t *Tunnel) Stop() {
	logger.Log.Infow("stopping tunnel", "tunnel", t.Config.Name)
	t.mu.Lock()
	if t.cancel != nil {
		t.cancel()
	}
	// Close listener to unblock Accept()
	if t.listener != nil {
		t.listener.Close()
	}
	t.mu.Unlock()
}

func (t *Tunnel) run() {
	retries := 0
	for {
		select {
		case <-t.ctx.Done():
			t.cleanup()
			t.setState(StateStopped, "")
			return
		default:
		}

		t.setState(StateConnecting, "")
		err := t.connect()
		if err != nil {
			select {
			case <-t.ctx.Done():
				t.cleanup()
				t.setState(StateStopped, "")
				return
			default:
			}

			retries++
			if retries >= maxRetries {
				logger.Log.Errorw("max retries reached, stopping tunnel",
					"tunnel", t.Config.Name,
					"retries", retries,
					"error", err,
				)
				t.cleanup()
				t.setState(StateError, fmt.Sprintf("max retries (%d) reached: %s", maxRetries, err.Error()))
				return
			}

			logger.Log.Warnw("connection failed, will retry",
				"tunnel", t.Config.Name,
				"error", err,
				"retry", retries,
				"max_retries", maxRetries,
				"retry_in", t.reconnectInterval,
			)
			t.setState(StateReconnecting, err.Error())
			time.Sleep(t.reconnectInterval)
			continue
		}

		// Connected successfully, reset retry counter
		retries = 0
		t.setState(StateConnected, "")
		t.serve()

		select {
		case <-t.ctx.Done():
			t.cleanup()
			t.setState(StateStopped, "")
			return
		default:
			logger.Log.Warnw("connection lost, reconnecting",
				"tunnel", t.Config.Name,
			)
			t.setState(StateReconnecting, "connection lost")
			time.Sleep(t.reconnectInterval)
		}
	}
}

func (t *Tunnel) connect() error {
	// Stop previous keepalive if any
	t.mu.Lock()
	if t.keepaliveCancel != nil {
		logger.Log.Debugw("stopping previous keepalive", "tunnel", t.Config.Name)
		t.keepaliveCancel()
		t.keepaliveCancel = nil
	}
	t.mu.Unlock()

	authMethods, err := t.getAuthMethods()
	if err != nil {
		return fmt.Errorf("auth setup failed: %w", err)
	}

	logger.Log.Debugw("connecting to SSH server",
		"tunnel", t.Config.Name,
		"host", t.Config.SSHAddr(),
		"user", t.Config.User,
		"auth_method", t.Config.AuthMethod,
	)

	config := &ssh.ClientConfig{
		User:            t.Config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", t.Config.SSHAddr(), config)
	if err != nil {
		return fmt.Errorf("ssh dial failed: %w", err)
	}
	t.client = client

	logger.Log.Debugw("SSH connected, setting up remote listener",
		"tunnel", t.Config.Name,
		"remote_addr", t.Config.RemoteAddr(),
	)

	listener, err := client.Listen("tcp", t.Config.RemoteAddr())
	if err != nil {
		client.Close()
		return fmt.Errorf("remote listen failed: %w", err)
	}
	t.listener = listener

	logger.Log.Infow("tunnel established",
		"tunnel", t.Config.Name,
		"remote", t.Config.RemoteAddr(),
		"local", t.Config.LocalAddr(),
	)

	// Start keepalive goroutine with its own cancel context
	keepaliveCtx, keepaliveCancel := context.WithCancel(t.ctx)
	t.mu.Lock()
	t.keepaliveCancel = keepaliveCancel
	t.mu.Unlock()
	logger.Log.Debugw("starting keepalive", "tunnel", t.Config.Name, "interval", keepaliveInterval)
	go t.keepalive(keepaliveCtx)

	return nil
}

func (t *Tunnel) getAuthMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	logger.Log.Debugw("getting auth methods", "tunnel", t.Config.Name, "auth_method", t.Config.AuthMethod)

	switch t.Config.AuthMethod {
	case "key":
		keyPath := t.Config.KeyPath
		if keyPath == "" {
			home, _ := os.UserHomeDir()
			keyPath = home + "/.ssh/id_rsa"
			logger.Log.Debugw("using default key path", "path", keyPath)
		} else if len(keyPath) > 0 && keyPath[0] == '~' {
			home, _ := os.UserHomeDir()
			keyPath = home + keyPath[1:]
		}
		logger.Log.Debugw("loading SSH key", "tunnel", t.Config.Name, "path", keyPath)
		key, err := os.ReadFile(keyPath)
		if err != nil {
			logger.Log.Errorw("failed to read key file", "tunnel", t.Config.Name, "path", keyPath, "error", err)
			return nil, fmt.Errorf("read key file: %w", err)
		}
		logger.Log.Debugw("key file loaded", "tunnel", t.Config.Name, "size", len(key))
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			logger.Log.Errorw("failed to parse private key", "tunnel", t.Config.Name, "error", err)
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		logger.Log.Debugw("SSH key parsed successfully", "tunnel", t.Config.Name)
		methods = append(methods, ssh.PublicKeys(signer))

	case "password":
		logger.Log.Debugw("using password auth", "tunnel", t.Config.Name)
		methods = append(methods, ssh.Password(t.Config.Password))

	default:
		logger.Log.Errorw("unknown auth method", "tunnel", t.Config.Name, "auth_method", t.Config.AuthMethod)
		return nil, fmt.Errorf("unknown auth method: %s", t.Config.AuthMethod)
	}

	logger.Log.Debugw("auth methods ready", "tunnel", t.Config.Name, "method_count", len(methods))
	return methods, nil
}

func (t *Tunnel) serve() {
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		conn, err := t.listener.Accept()
		if err != nil {
			logger.Log.Debugw("listener accept error", "tunnel", t.Config.Name, "error", err)
			return
		}

		logger.Log.Debugw("new connection",
			"tunnel", t.Config.Name,
			"remote_client", conn.RemoteAddr().String(),
		)

		go t.handleConnection(conn)
	}
}

func (t *Tunnel) handleConnection(remoteConn net.Conn) {
	defer remoteConn.Close()

	localConn, err := net.Dial("tcp", t.Config.LocalAddr())
	if err != nil {
		logger.Log.Warnw("failed to connect to local service",
			"tunnel", t.Config.Name,
			"local_addr", t.Config.LocalAddr(),
			"error", err,
		)
		return
	}
	defer localConn.Close()

	logger.Log.Debugw("forwarding connection",
		"tunnel", t.Config.Name,
		"from", remoteConn.RemoteAddr().String(),
		"to", t.Config.LocalAddr(),
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(localConn, remoteConn)
	}()

	go func() {
		defer wg.Done()
		io.Copy(remoteConn, localConn)
	}()

	wg.Wait()

	logger.Log.Debugw("connection closed",
		"tunnel", t.Config.Name,
		"remote_client", remoteConn.RemoteAddr().String(),
	)
}

func (t *Tunnel) cleanup() {
	t.mu.Lock()
	if t.keepaliveCancel != nil {
		t.keepaliveCancel()
		t.keepaliveCancel = nil
	}
	t.mu.Unlock()

	if t.listener != nil {
		t.listener.Close()
		t.listener = nil
	}
	if t.client != nil {
		t.client.Close()
		t.client = nil
	}
}

func (t *Tunnel) keepalive(ctx context.Context) {
	ticker := time.NewTicker(keepaliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Debugw("keepalive stopped", "tunnel", t.Config.Name)
			return
		case <-ticker.C:
			t.mu.RLock()
			client := t.client
			t.mu.RUnlock()

			if client == nil {
				return
			}

			_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
			if err != nil {
				logger.Log.Warnw("keepalive failed, connection may be dead",
					"tunnel", t.Config.Name,
					"error", err,
				)
				// Close listener to trigger reconnect
				t.mu.Lock()
				if t.listener != nil {
					t.listener.Close()
				}
				t.mu.Unlock()
				return
			}
			logger.Log.Infow("keepalive ok", "tunnel", t.Config.Name)
		}
	}
}
