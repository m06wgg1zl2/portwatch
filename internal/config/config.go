package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config holds the top-level portwatch configuration.
type Config struct {
	Interval time.Duration `json:"-"`
	IntervalSeconds int        `json:"interval_seconds"`
	Ports    []PortConfig  `json:"ports"`
}

// PortConfig describes a single port to monitor and its callbacks.
type PortConfig struct {
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Webhooks []string `json:"webhooks,omitempty"`
	Shell    string   `json:"shell,omitempty"`
}

// Load reads and parses a JSON config file from the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: decode: %w", err)
	}

	if cfg.IntervalSeconds <= 0 {
		cfg.IntervalSeconds = 30
	}
	cfg.Interval = time.Duration(cfg.IntervalSeconds) * time.Second

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	for i, p := range c.Ports {
		if p.Host == "" {
			return fmt.Errorf("config: port[%d]: host is required", i)
		}
		if p.Port <= 0 || p.Port > 65535 {
			return fmt.Errorf("config: port[%d]: invalid port number %d", i, p.Port)
		}
	}
	return nil
}
