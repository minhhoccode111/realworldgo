package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadWithEnvOverrides(t *testing.T) {
	t.Setenv("SERVER_HOST", "0.0.0.0")
	t.Setenv("SERVER_PORT", "8081")
	t.Setenv("SERVER_READ_TIMEOUT", "5s")
	t.Setenv("SERVER_WRITE_TIMEOUT", "6s")
	t.Setenv("SERVER_IDLE_TIMEOUT", "30s")
	t.Setenv("DB_NAME", "rw_test")
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USERNAME", "tester")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_SSL_MODE", "require")
	t.Setenv("DB_SCHEMA", "tenant1")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_EXPIRATION", "24h")
	t.Setenv("JWT_ISSUER", "suite")
	t.Setenv("ACCESS_CONTROL_ALLOW_ORIGIN", "https://app.example.com, https://api.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected load to succeed, got error: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("unexpected server host: %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8081 {
		t.Fatalf("unexpected server port: %d", cfg.Server.Port)
	}
	if cfg.Server.ReadTimeout != 5*time.Second {
		t.Fatalf("unexpected read timeout: %s", cfg.Server.ReadTimeout)
	}
	if cfg.Database.DatabaseURL() != "postgres://tester:secret@db.local:5433/rw_test?sslmode=require&search_path=tenant1" {
		t.Fatalf("unexpected database url: %s", cfg.Database.DatabaseURL())
	}

	if cfg.JWT.Secret != "test-secret" {
		t.Fatalf("unexpected jwt secret: %s", cfg.JWT.Secret)
	}
	if cfg.JWT.Expiration != 24*time.Hour {
		t.Fatalf("unexpected jwt expiration: %s", cfg.JWT.Expiration)
	}
	if cfg.JWT.Issuer != "suite" {
		t.Fatalf("unexpected jwt issuer: %s", cfg.JWT.Issuer)
	}

	expectedOrigins := []string{"https://app.example.com", "https://api.example.com"}
	if len(cfg.CORS.AllowedOrigins) != len(expectedOrigins) {
		t.Fatalf("unexpected origins length: %v", cfg.CORS.AllowedOrigins)
	}
	for i, origin := range expectedOrigins {
		if cfg.CORS.AllowedOrigins[i] != origin {
			t.Fatalf("unexpected origin at %d: %s", i, cfg.CORS.AllowedOrigins[i])
		}
	}
}

func TestLoadValidationErrors(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("SERVER_PORT", "70000")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("ACCESS_CONTROL_ALLOW_ORIGIN", "*")

	if _, err := Load(); err == nil {
		t.Fatal("expected validation error for empty jwt secret")
	}

	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("DB_PASSWORD", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected validation error for empty db password")
	}

	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("SERVER_PORT", "70000")
	if _, err := Load(); err == nil {
		t.Fatal("expected validation error for invalid server port")
	}

	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("DB_PORT", "70000")
	if _, err := Load(); err == nil {
		t.Fatal("expected validation error for invalid db port")
	}

	t.Setenv("DB_PORT", "5432")
	t.Setenv("ACCESS_CONTROL_ALLOW_ORIGIN", "ftp://invalid.example.com")
	if _, err := Load(); err == nil {
		t.Fatal("expected validation error for invalid origin scheme")
	}

	// restore to avoid leaking env mutations in other tests that might read host defaults
	os.Unsetenv("SERVER_PORT")
}
