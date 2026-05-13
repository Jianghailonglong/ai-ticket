package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Ticket TicketConfig `yaml:"ticket"`
	Auth   AuthConfig   `yaml:"auth"`
}

type ServerConfig struct {
	Transport string `yaml:"transport"`
	HTTPAddr  string `yaml:"http_addr"`
}

type TicketConfig struct {
	BaseURL string `yaml:"base_url"`
	Timeout int    `yaml:"timeout"`
}

type AuthConfig struct {
	TokenHeader string `yaml:"token_header"`
}

func (t *TicketConfig) GetTimeout() time.Duration {
	if t.Timeout <= 0 {
		return 30 * time.Second
	}
	return time.Duration(t.Timeout) * time.Second
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Server: ServerConfig{
			Transport: "stdio",
			HTTPAddr:  ":8080",
		},
		Auth: AuthConfig{
			TokenHeader: "Authorization",
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
