package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/aleeeeeexxxx/ssht/internal/tunnel"
)

type Config struct {
	Tunnels []tunnel.Config `json:"tunnels"`
}

func DefaultPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "ssht", "config.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ssht", "config.json")
}

func StatePath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "ssht", "state.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "ssht", "state.json")
}

func KeysDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "ssht", "keys")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "ssht", "keys")
}

func SaveKey(tunnelName, content string) (string, error) {
	dir := KeysDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	keyPath := filepath.Join(dir, tunnelName+".key")
	if err := os.WriteFile(keyPath, []byte(content), 0600); err != nil {
		return "", err
	}

	return keyPath, nil
}

func DeleteKey(tunnelName string) error {
	keyPath := filepath.Join(KeysDir(), tunnelName+".key")
	if err := os.Remove(keyPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Tunnels: []tunnel.Config{}}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	for i := range cfg.Tunnels {
		if cfg.Tunnels[i].Port == 0 {
			cfg.Tunnels[i].Port = 22
		}
		if cfg.Tunnels[i].LocalHost == "" {
			cfg.Tunnels[i].LocalHost = "127.0.0.1"
		}
		if cfg.Tunnels[i].RemoteHost == "" {
			cfg.Tunnels[i].RemoteHost = "0.0.0.0"
		}
	}

	return &cfg, nil
}

func Save(path string, cfg *Config) error {
	if path == "" {
		path = DefaultPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (c *Config) Add(tc tunnel.Config) {
	for i, t := range c.Tunnels {
		if t.Name == tc.Name {
			c.Tunnels[i] = tc
			return
		}
	}
	c.Tunnels = append(c.Tunnels, tc)
}

func (c *Config) Remove(name string) {
	for i, t := range c.Tunnels {
		if t.Name == name {
			c.Tunnels = append(c.Tunnels[:i], c.Tunnels[i+1:]...)
			return
		}
	}
}

func (c *Config) Get(name string) *tunnel.Config {
	for i := range c.Tunnels {
		if c.Tunnels[i].Name == name {
			return &c.Tunnels[i]
		}
	}
	return nil
}

// LoadWithMerge loads state, merges config into it, and saves back to state.
// Merge logic:
//   - config has, state doesn't → add
//   - config has, state has (same name) → config overwrites
//   - config doesn't have, state has → keep
func LoadWithMerge(configPath, statePath string) (*Config, error) {
	if configPath == "" {
		configPath = DefaultPath()
	}
	if statePath == "" {
		statePath = StatePath()
	}

	// Load state first
	state, _ := Load(statePath)
	if state == nil {
		state = &Config{Tunnels: []tunnel.Config{}}
	}

	// Load config
	cfg, _ := Load(configPath)
	if cfg == nil || len(cfg.Tunnels) == 0 {
		// No config to merge, just return state
		return state, nil
	}

	// Build a map of state tunnels by name
	stateMap := make(map[string]int)
	for i, t := range state.Tunnels {
		stateMap[t.Name] = i
	}

	// Merge config into state
	for _, ct := range cfg.Tunnels {
		if idx, exists := stateMap[ct.Name]; exists {
			// Config overwrites state
			state.Tunnels[idx] = ct
		} else {
			// Add new from config
			state.Tunnels = append(state.Tunnels, ct)
		}
	}

	// Apply defaults
	for i := range state.Tunnels {
		if state.Tunnels[i].Port == 0 {
			state.Tunnels[i].Port = 22
		}
		if state.Tunnels[i].LocalHost == "" {
			state.Tunnels[i].LocalHost = "127.0.0.1"
		}
		if state.Tunnels[i].RemoteHost == "" {
			state.Tunnels[i].RemoteHost = "0.0.0.0"
		}
	}

	// Save merged result to state
	if err := Save(statePath, state); err != nil {
		return nil, err
	}

	return state, nil
}
