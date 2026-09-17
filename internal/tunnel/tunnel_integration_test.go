//go:build integration

package tunnel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/aleeeeeexxxx/ssht/internal/logger"
)

const testTunnelName = "integration-test"
const reconnectTunnelName = "reconnect-test"

// Integration test config - set via environment variables or use defaults
type testConfig struct {
	SSHHost    string
	SSHPort    int
	SSHUser    string
	SSHKeyPath string
	RemotePort int
	LocalPort  int
}

func getTestConfig() testConfig {
	cfg := testConfig{
		SSHHost:    "10.237.153.61",
		SSHPort:    22,
		SSHUser:    "ccloud",
		SSHKeyPath: os.Getenv("HOME") + "/.ssh/another_vm",
		RemotePort: 19999,
		LocalPort:  18888,
	}

	if h := os.Getenv("TEST_SSH_HOST"); h != "" {
		cfg.SSHHost = h
	}
	if u := os.Getenv("TEST_SSH_USER"); u != "" {
		cfg.SSHUser = u
	}
	if k := os.Getenv("TEST_SSH_KEY"); k != "" {
		cfg.SSHKeyPath = k
	}

	return cfg
}

func TestTunnelIntegration(t *testing.T) {
	logger.Init(true)
	defer logger.Sync()

	cfg := getTestConfig()

	// 1. Start local HTTP server
	localServer := startLocalServer(t, cfg.LocalPort)
	defer localServer.Close()

	// 2. Create and start tunnel
	manager := NewManager()
	tunnelCfg := Config{
		Name:       testTunnelName,
		Host:       cfg.SSHHost,
		Port:       cfg.SSHPort,
		User:       cfg.SSHUser,
		AuthMethod: "key",
		KeyPath:    cfg.SSHKeyPath,
		RemoteHost: "127.0.0.1",
		RemotePort: cfg.RemotePort,
		LocalHost:  "127.0.0.1",
		LocalPort:  cfg.LocalPort,
	}

	// Check initial state is Stopped
	manager.Add(tunnelCfg)
	if state := manager.GetState(testTunnelName); state != StateStopped {
		t.Errorf("expected initial state Stopped, got %v", state)
	}

	if err := manager.Start(testTunnelName); err != nil {
		t.Fatalf("failed to start tunnel: %v", err)
	}
	defer manager.StopAll()

	// 3. Wait for tunnel to connect and verify state transitions
	if !waitForState(manager, testTunnelName, StateConnected, 10*time.Second) {
		t.Fatalf("tunnel did not connect in time, state: %v", manager.GetState(testTunnelName))
	}

	// Verify connected state
	if state := manager.GetState(testTunnelName); state != StateConnected {
		t.Errorf("expected state Connected, got %v", state)
	}

	// Get tunnel and check its state directly
	tun := manager.Get(testTunnelName)
	if tun == nil {
		t.Fatal("tunnel not found in manager")
	}
	if tun.State() != StateConnected {
		t.Errorf("tunnel.State() expected Connected, got %v", tun.State())
	}
	if tun.ErrorMsg() != "" {
		t.Errorf("tunnel.ErrorMsg() expected empty, got %q", tun.ErrorMsg())
	}

	// 4. Test via SSH remote curl
	resp, err := remoteCurl(cfg, cfg.RemotePort)
	if err != nil {
		t.Fatalf("remote curl failed: %v", err)
	}

	// 5. Verify response
	if !strings.Contains(resp, "integration-test-ok") {
		t.Errorf("unexpected response: %s", resp)
	}

	// 6. Stop tunnel and verify state
	manager.Stop(testTunnelName)
	time.Sleep(100 * time.Millisecond) // Give goroutine time to process cancel
	if !waitForState(manager, testTunnelName, StateStopped, 5*time.Second) {
		t.Errorf("tunnel did not stop in time, state: %v", manager.GetState(testTunnelName))
	}

	t.Logf("Integration test passed! Response: %s", resp)
}

func TestTunnelReconnect(t *testing.T) {
	logger.Init(true)
	defer logger.Sync()

	cfg := getTestConfig()

	// Start local server
	localServer := startLocalServer(t, cfg.LocalPort+1)
	defer localServer.Close()

	// Create tunnel
	manager := NewManager()
	tunnelCfg := Config{
		Name:       reconnectTunnelName,
		Host:       cfg.SSHHost,
		Port:       cfg.SSHPort,
		User:       cfg.SSHUser,
		AuthMethod: "key",
		KeyPath:    cfg.SSHKeyPath,
		RemoteHost: "127.0.0.1",
		RemotePort: cfg.RemotePort + 1,
		LocalHost:  "127.0.0.1",
		LocalPort:  cfg.LocalPort + 1,
	}

	manager.Add(tunnelCfg)
	if err := manager.Start(reconnectTunnelName); err != nil {
		t.Fatalf("failed to start tunnel: %v", err)
	}
	defer manager.StopAll()

	// Wait for connection
	if !waitForState(manager, reconnectTunnelName, StateConnected, 10*time.Second) {
		t.Fatalf("tunnel did not connect")
	}

	// First request should work
	resp, err := remoteCurl(cfg, cfg.RemotePort+1)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	if !strings.Contains(resp, "integration-test-ok") {
		t.Errorf("unexpected response: %s", resp)
	}

	t.Log("First request succeeded, tunnel is working")
}

func startLocalServer(t *testing.T, port int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"status":  "integration-test-ok",
			"message": "Hello from local server",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)
	return server
}

func waitForState(manager *Manager, name string, target State, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if manager.GetState(name) == target {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func remoteCurl(cfg testConfig, port int) (string, error) {
	cmd := exec.Command("ssh",
		"-i", cfg.SSHKeyPath,
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=5",
		fmt.Sprintf("%s@%s", cfg.SSHUser, cfg.SSHHost),
		fmt.Sprintf("curl -s localhost:%d", port),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ssh curl failed: %w, output: %s", err, output)
	}

	return string(output), nil
}
