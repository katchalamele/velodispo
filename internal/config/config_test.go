package config

import (
	"os"
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"HTTP_ADDR", "SOURCES", "POLL_INTERVAL", "HTTP_TIMEOUT", "POLL_MAX_CONCURRENCY",
	} {
		if v, ok := os.LookupEnv(k); ok {
			key, val := k, v
			os.Unsetenv(key)
			t.Cleanup(func() { os.Setenv(key, val) })
		}
	}
}

func TestDSN(t *testing.T) {
	db := DBConfig{
		Host:     "db",
		Port:     5432,
		User:     "u",
		Password: "p",
		Name:     "velodispo",
		SSLMode:  "disable",
	}
	want := "postgres://u:p@db:5432/velodispo?sslmode=disable"
	if got := db.DSN(); got != want {
		t.Errorf("DSN() = %q, attendu %q", got, want)
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DB.Host != "db" || cfg.DB.Port != 5432 || cfg.DB.Name != "velodispo" {
		t.Errorf("valeurs par défaut inattendues: %+v", cfg.DB)
	}
}

func TestLoadEnvOverride(t *testing.T) {
	t.Setenv("DB_HOST", "example")
	t.Setenv("DB_PORT", "6543")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DB.Host != "example" || cfg.DB.Port != 6543 {
		t.Errorf("surcharge env non prise en compte: %+v", cfg.DB)
	}
}
