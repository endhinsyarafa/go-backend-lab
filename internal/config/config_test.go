package config_test

import (
	"strings"
	"testing"

	"github.com/latihan/go-backend-lab/internal/config"
)

func TestConfigLoadAndDSN(t *testing.T) {
	t.Setenv("APP_PORT", "9090")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "test_db")
	t.Setenv("DB_SSLMODE", "disable")
	t.Setenv("JWT_SECRET", "test_secret_key")

	cfg := config.Load()

	if cfg.JWTSecret != "test_secret_key" {
		t.Errorf("expected JWTSecret 'test_secret_key', got '%s'", cfg.JWTSecret)
	}

	dsn := cfg.DSN()
	if !strings.Contains(dsn, "test_db") || !strings.Contains(dsn, "postgres") {
		t.Errorf("DSN tidak sesuai harapan: %s", dsn)
	}
}
