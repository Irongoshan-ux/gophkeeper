// Package config loads server and client runtime settings.
package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// ServerConfig holds gRPC server settings.
type ServerConfig struct {
	Address     string `env:"SERVER_ADDRESS"`
	DatabaseDSN string `env:"DATABASE_DSN"`
	JWTSecret   string `env:"JWT_SECRET"`
	EnableTLS   bool   `env:"ENABLE_TLS"`
}

// DefaultServer returns default server configuration.
func DefaultServer() ServerConfig {
	return ServerConfig{
		Address:   "localhost:9090",
		JWTSecret: "default-secret",
	}
}

// LoadServer parses server flags and environment variables.
func LoadServer() (*ServerConfig, error) {
	cfg := DefaultServer()
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.StringVar(&cfg.Address, "a", cfg.Address, "gRPC server address")
	fs.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "PostgreSQL DSN")
	fs.StringVar(&cfg.JWTSecret, "s", cfg.JWTSecret, "JWT signing secret")
	fs.BoolVar(&cfg.EnableTLS, "tls", cfg.EnableTLS, "enable TLS")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, err
	}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ClientConfig holds CLI client settings.
type ClientConfig struct {
	ServerAddress string `json:"server_address"`
	Token         string `json:"token,omitempty"`
	Salt          []byte `json:"salt,omitempty"`
}
