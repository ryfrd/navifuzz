package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server   string `toml:"server"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Scrobble bool   `toml:"scrobble"`
	Selector string `toml:"selector"`
	Player   string `toml:"player"`
	Shuffle  bool   `toml:"shuffle"`
	Loop     bool   `toml:"loop"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find home directory: %w", err)
	}

	path := filepath.Join(home, ".config", "navifzf", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config %s: %w", path, err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}

	if cfg.Server == "" {
		return nil, fmt.Errorf("config: server is required")
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("config: username is required")
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("config: password is required")
	}

	if cfg.Selector == "" {
		cfg.Selector = "fzf"
	}
	if cfg.Player == "" {
		cfg.Player = "mpv"
	}

	return &cfg, nil
}
