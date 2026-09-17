package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/aleeeeeexxxx/ssht/internal/tunnel"
)

type Config struct {
	Tunnels []tunnel.Config `json:"tunnels"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ssht", "config.json")
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
