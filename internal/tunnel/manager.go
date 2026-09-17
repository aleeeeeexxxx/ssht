package tunnel

import (
	"sync"

	"github.com/aleeeeeexxxx/ssht/internal/logger"
)

type Manager struct {
	tunnels      map[string]*Tunnel
	mu           sync.RWMutex
	stateChanges chan StateChange
}

func NewManager() *Manager {
	logger.Log.Debug("creating tunnel manager")
	return &Manager{
		tunnels:      make(map[string]*Tunnel),
		stateChanges: make(chan StateChange, 100),
	}
}

func (m *Manager) StateChanges() <-chan StateChange {
	return m.stateChanges
}

func (m *Manager) Add(cfg Config) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tunnels[cfg.Name]; exists {
		logger.Log.Infow("replacing existing tunnel", "tunnel", cfg.Name)
		m.tunnels[cfg.Name].Stop()
	}
	m.tunnels[cfg.Name] = New(cfg, m.stateChanges)
	logger.Log.Infow("tunnel added", "tunnel", cfg.Name)
}

func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if t, exists := m.tunnels[name]; exists {
		t.Stop()
		delete(m.tunnels, name)
		logger.Log.Infow("tunnel removed", "tunnel", name)
	}
}

func (m *Manager) Start(name string) error {
	m.mu.RLock()
	t, exists := m.tunnels[name]
	m.mu.RUnlock()

	if !exists {
		logger.Log.Warnw("tunnel not found", "tunnel", name)
		return nil
	}
	return t.Start()
}

func (m *Manager) Stop(name string) {
	m.mu.RLock()
	t, exists := m.tunnels[name]
	m.mu.RUnlock()

	if exists {
		t.Stop()
	}
}

func (m *Manager) StopAll() {
	logger.Log.Info("stopping all tunnels")
	m.mu.RLock()
	tunnels := make([]*Tunnel, 0, len(m.tunnels))
	for _, t := range m.tunnels {
		tunnels = append(tunnels, t)
	}
	m.mu.RUnlock()

	for _, t := range tunnels {
		t.Stop()
	}
}

func (m *Manager) Get(name string) *Tunnel {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tunnels[name]
}

func (m *Manager) List() []*Tunnel {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Tunnel, 0, len(m.tunnels))
	for _, t := range m.tunnels {
		result = append(result, t)
	}
	return result
}

func (m *Manager) GetState(name string) State {
	m.mu.RLock()
	t, exists := m.tunnels[name]
	m.mu.RUnlock()

	if !exists {
		return StateStopped
	}
	return t.State()
}
