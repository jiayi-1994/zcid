package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Redis      RedisConfig      `yaml:"redis"`
	MinIO      MinIOConfig      `yaml:"minio"`
	Auth       AuthConfig       `yaml:"auth"`
	Encryption EncryptionConfig `yaml:"-"`
	K8s        K8sConfig        `yaml:"-"`
	ArgoCD     ArgoCDConfig     `yaml:"-"`
}

type K8sConfig struct {
	Enabled    bool   `yaml:"-"`
	Namespace  string `yaml:"-"`
}

type ArgoCDConfig struct {
	Enabled  bool   `yaml:"-"`
	Server   string `yaml:"-"`
	Token    string `yaml:"-"`
	Insecure bool   `yaml:"-"`
}

type ServerConfig struct {
	Port     string `yaml:"port"`
	LogLevel string `yaml:"log_level"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type MinIOConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	UseSSL    bool   `yaml:"use_ssl"`
}

type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

type EncryptionConfig struct {
	Key string `yaml:"-"` // only from env var, never from YAML
}

// Load reads config.yaml and applies environment variable overrides.
// Environment variables take precedence over YAML values.
// Sensitive fields (passwords, keys) are only loaded from environment variables.
func Load(path string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:     "8080",
			LogLevel: "info",
		},
		Database: DatabaseConfig{
			Host:    "localhost",
			Port:    5432,
			Name:    "zcid",
			User:    "zcid",
			SSLMode: "disable",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		MinIO: MinIOConfig{
			Endpoint: "localhost:9000",
			UseSSL:   false,
		},
		Auth: AuthConfig{},
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		// config.yaml not found is acceptable; use defaults + env
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config file: %w", err)
		}
	}

	applyEnvOverrides(cfg)

	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	// Server
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("SERVER_LOG_LEVEL"); v != "" {
		cfg.Server.LogLevel = v
	}

	// Database
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Database.Port = port
		}
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Name = v
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_SSL_MODE"); v != "" {
		cfg.Database.SSLMode = v
	}
	// Password: env overrides yaml
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}

	// Redis
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Redis.Port = port
		}
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		if db, err := strconv.Atoi(v); err == nil {
			cfg.Redis.DB = db
		}
	}
	// Password: env overrides yaml
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}

	// MinIO
	if v := os.Getenv("MINIO_ENDPOINT"); v != "" {
		cfg.MinIO.Endpoint = v
	}
	if v := os.Getenv("MINIO_ACCESS_KEY"); v != "" {
		cfg.MinIO.AccessKey = v
	}
	if v := os.Getenv("MINIO_USE_SSL"); v != "" {
		cfg.MinIO.UseSSL = v == "true"
	}
	// Secret key: env overrides yaml
	if v := os.Getenv("MINIO_SECRET_KEY"); v != "" {
		cfg.MinIO.SecretKey = v
	}

	// Auth
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}

	// Encryption
	if v := os.Getenv("ZCID_ENCRYPTION_KEY"); v != "" {
		cfg.Encryption.Key = v
	}

	// K8s
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" || os.Getenv("KUBECONFIG") != "" {
		cfg.K8s.Enabled = true
	}
	if v := os.Getenv("ZCID_K8S_ENABLED"); v == "true" {
		cfg.K8s.Enabled = true
	} else if v == "false" {
		cfg.K8s.Enabled = false
	}
	cfg.K8s.Namespace = os.Getenv("ZCID_K8S_NAMESPACE")
	if cfg.K8s.Namespace == "" {
		cfg.K8s.Namespace = "zcicd"
	}

	// ArgoCD
	if v := os.Getenv("ARGOCD_SERVER"); v != "" {
		cfg.ArgoCD.Server = v
		cfg.ArgoCD.Enabled = true
	}
	if v := os.Getenv("ARGOCD_AUTH_TOKEN"); v != "" {
		cfg.ArgoCD.Token = v
	}
	if os.Getenv("ARGOCD_INSECURE") == "true" {
		cfg.ArgoCD.Insecure = true
	}
}

// DSN returns the PostgreSQL connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// MigrationURL returns the PostgreSQL URL format required by golang-migrate.
func (c *DatabaseConfig) MigrationURL() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.Name,
	}

	q := u.Query()
	if c.SSLMode != "" {
		q.Set("sslmode", c.SSLMode)
	}
	u.RawQuery = q.Encode()

	return u.String()
}
