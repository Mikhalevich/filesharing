package service

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Configer interface {
	Service() Config
	Validate() error
}

type Config struct {
	Port int `yaml:"port"`
}

func (c Config) Validate() error {
	if c.Port <= 0 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	return nil
}

func loadConfig(cfg Configer, path string) error {
	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file error: %w", err)
	}

	if err := yaml.Unmarshal(buf, cfg); err != nil {
		return fmt.Errorf("unmarshal yml error: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate error: %w", err)
	}

	if err := cfg.Service().Validate(); err != nil {
		return fmt.Errorf("service validate error: %w", err)
	}

	return nil
}
