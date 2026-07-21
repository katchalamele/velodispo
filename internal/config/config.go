package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DB       DBConfig
	Poll     PollConfig
	HTTPAddr string `envconfig:"HTTP_ADDR" default:":8081"`
}

type PollConfig struct {
	Interval       time.Duration `envconfig:"POLL_INTERVAL" default:"60s"`
	Sources        []string      `envconfig:"SOURCES" default:"nantes,paris"`
	HTTPTimeout    time.Duration `envconfig:"HTTP_TIMEOUT" default:"15s"`
	MaxConcurrency int           `envconfig:"POLL_MAX_CONCURRENCY" default:"4"`
}

type DBConfig struct {
	Host     string `envconfig:"DB_HOST" default:"db"`
	Port     int    `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"velodispo"`
	Password string `envconfig:"DB_PASSWORD" default:"velodispo"`
	Name     string `envconfig:"DB_NAME" default:"velodispo"`
	SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return Config{}, fmt.Errorf("chargement config: %w", err)
	}
	return cfg, nil
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode,
	)
}
