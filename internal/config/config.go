// Package config handles tracker configuration and state transitions.
package config

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	User   string `env:"WORK_USER"`
	Editor string `env:"EDITOR"`
	Visual string `env:"VISUAL"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, fmt.Errorf("parsing env config: %w", err)
	}
	if cfg.Editor == "" {
		if cfg.Visual != "" {
			cfg.Editor = cfg.Visual
		} else {
			cfg.Editor = "vi"
		}
	}
	if cfg.User == "" {
		cfg.User = resolveGitUser()
	}
	return cfg, nil
}

func resolveGitUser() string {
	out, err := exec.Command("git", "config", "user.name").Output()
	if err == nil {
		if name := strings.TrimSpace(string(out)); name != "" {
			return name
		}
	}
	return "system"
}
