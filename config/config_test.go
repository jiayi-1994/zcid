package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Load with a non-existent file should use defaults
	cfg, err := Load("nonexistent.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("expected default port '8080', got '%s'", cfg.Server.Port)
	}
	if cfg.Server.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got '%s'", cfg.Server.LogLevel)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("expected default db port 5432, got %d", cfg.Database.Port)
	}
	if cfg.Database.Name != "zcid" {
		t.Errorf("expected default db name 'zcid', got '%s'", cfg.Database.Name)
	}
	if cfg.Redis.Host != "localhost" {
		t.Errorf("expected default redis host 'localhost', got '%s'", cfg.Redis.Host)
	}
	if cfg.Redis.Port != 6379 {
		t.Errorf("expected default redis port 6379, got %d", cfg.Redis.Port)
	}
	if cfg.MinIO.Endpoint != "localhost:9000" {
		t.Errorf("expected default minio endpoint 'localhost:9000', got '%s'", cfg.MinIO.Endpoint)
	}
	if cfg.Auth.JWTSecret != "" {
		t.Errorf("expected default jwt secret empty, got '%s'", cfg.Auth.JWTSecret)
	}
}

func TestLoad_YAMLOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")

	yamlContent := []byte(`
server:
  port: "9090"
  log_level: debug
database:
  host: db.example.com
  port: 5433
  name: testdb
  user: testuser
  ssl_mode: require
redis:
  host: redis.example.com
  port: 6380
  db: 1
minio:
  endpoint: minio.example.com:9000
  access_key: testaccesskey
  use_ssl: true
`)
	if err := os.WriteFile(yamlPath, yamlContent, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Server.Port != "9090" {
		t.Errorf("expected port '9090', got '%s'", cfg.Server.Port)
	}
	if cfg.Server.LogLevel != "debug" {
		t.Errorf("expected log level 'debug', got '%s'", cfg.Server.LogLevel)
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("expected db host 'db.example.com', got '%s'", cfg.Database.Host)
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("expected db port 5433, got %d", cfg.Database.Port)
	}
	if cfg.Database.Name != "testdb" {
		t.Errorf("expected db name 'testdb', got '%s'", cfg.Database.Name)
	}
	if cfg.Redis.Host != "redis.example.com" {
		t.Errorf("expected redis host 'redis.example.com', got '%s'", cfg.Redis.Host)
	}
	if cfg.Redis.Port != 6380 {
		t.Errorf("expected redis port 6380, got %d", cfg.Redis.Port)
	}
	if cfg.MinIO.Endpoint != "minio.example.com:9000" {
		t.Errorf("expected minio endpoint 'minio.example.com:9000', got '%s'", cfg.MinIO.Endpoint)
	}
	if !cfg.MinIO.UseSSL {
		t.Error("expected minio use_ssl true")
	}
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")

	yamlContent := []byte(`
server:
  port: "9090"
database:
  host: db.example.com
  port: 5433
  name: yamldb
  user: yamluser
  ssl_mode: require
`)
	if err := os.WriteFile(yamlPath, yamlContent, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Set env vars that should override YAML
	t.Setenv("SERVER_PORT", "3000")
	t.Setenv("SERVER_LOG_LEVEL", "error")
	t.Setenv("DB_HOST", "env-db-host")
	t.Setenv("DB_PORT", "5555")
	t.Setenv("DB_NAME", "envdb")
	t.Setenv("DB_USER", "envuser")
	t.Setenv("DB_PASSWORD", "secret-password")
	t.Setenv("DB_SSL_MODE", "disable")
	t.Setenv("REDIS_HOST", "env-redis")
	t.Setenv("REDIS_PORT", "6381")
	t.Setenv("REDIS_PASSWORD", "redis-secret")
	t.Setenv("MINIO_ENDPOINT", "env-minio:9000")
	t.Setenv("MINIO_ACCESS_KEY", "env-access")
	t.Setenv("MINIO_SECRET_KEY", "env-secret")
	t.Setenv("MINIO_USE_SSL", "true")
	t.Setenv("JWT_SECRET", "jwt-env-secret")

	cfg, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify env overrides YAML
	if cfg.Server.Port != "3000" {
		t.Errorf("expected port '3000' from env, got '%s'", cfg.Server.Port)
	}
	if cfg.Server.LogLevel != "error" {
		t.Errorf("expected log level 'error' from env, got '%s'", cfg.Server.LogLevel)
	}
	if cfg.Database.Host != "env-db-host" {
		t.Errorf("expected db host 'env-db-host' from env, got '%s'", cfg.Database.Host)
	}
	if cfg.Database.Port != 5555 {
		t.Errorf("expected db port 5555 from env, got %d", cfg.Database.Port)
	}
	if cfg.Database.Name != "envdb" {
		t.Errorf("expected db name 'envdb' from env, got '%s'", cfg.Database.Name)
	}
	if cfg.Database.User != "envuser" {
		t.Errorf("expected db user 'envuser' from env, got '%s'", cfg.Database.User)
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("expected ssl_mode 'disable' from env, got '%s'", cfg.Database.SSLMode)
	}

	// Verify sensitive fields from env only
	if cfg.Database.Password != "secret-password" {
		t.Errorf("expected db password 'secret-password' from env, got '%s'", cfg.Database.Password)
	}
	if cfg.Redis.Password != "redis-secret" {
		t.Errorf("expected redis password 'redis-secret' from env, got '%s'", cfg.Redis.Password)
	}
	if cfg.MinIO.SecretKey != "env-secret" {
		t.Errorf("expected minio secret key 'env-secret' from env, got '%s'", cfg.MinIO.SecretKey)
	}

	// Verify other env overrides
	if cfg.Redis.Host != "env-redis" {
		t.Errorf("expected redis host 'env-redis' from env, got '%s'", cfg.Redis.Host)
	}
	if cfg.Redis.Port != 6381 {
		t.Errorf("expected redis port 6381 from env, got %d", cfg.Redis.Port)
	}
	if cfg.MinIO.Endpoint != "env-minio:9000" {
		t.Errorf("expected minio endpoint 'env-minio:9000' from env, got '%s'", cfg.MinIO.Endpoint)
	}
	if cfg.MinIO.AccessKey != "env-access" {
		t.Errorf("expected minio access key 'env-access' from env, got '%s'", cfg.MinIO.AccessKey)
	}
	if !cfg.MinIO.UseSSL {
		t.Error("expected minio use_ssl true from env")
	}
	if cfg.Auth.JWTSecret != "jwt-env-secret" {
		t.Errorf("expected jwt secret from env, got '%s'", cfg.Auth.JWTSecret)
	}
}

func TestLoad_SensitiveFieldsNotInYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")

	// YAML includes password fields — they should be ignored due to yaml:"-" tag
	yamlContent := []byte(`
database:
  host: localhost
redis:
  host: localhost
minio:
  endpoint: localhost:9000
`)
	if err := os.WriteFile(yamlPath, yamlContent, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Don't set env vars for sensitive fields
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("MINIO_SECRET_KEY", "")

	cfg, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Database.Password != "" {
		t.Errorf("expected empty db password, got '%s'", cfg.Database.Password)
	}
	if cfg.Redis.Password != "" {
		t.Errorf("expected empty redis password, got '%s'", cfg.Redis.Password)
	}
	if cfg.MinIO.SecretKey != "" {
		t.Errorf("expected empty minio secret key, got '%s'", cfg.MinIO.SecretKey)
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "zcid",
		Password: "secret",
		Name:     "zcid",
		SSLMode:  "disable",
	}

	expected := "host=localhost port=5432 user=zcid password=secret dbname=zcid sslmode=disable"
	if dsn := cfg.DSN(); dsn != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, dsn)
	}
}
