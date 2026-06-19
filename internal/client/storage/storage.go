// Package storage persists CLI configuration and local sync cache.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Irongoshan-ux/gophkeeper/internal/config"
	"github.com/Irongoshan-ux/gophkeeper/internal/model"
)

// ErrNotConfigured is returned when the client config file is missing.
var ErrNotConfigured = errors.New("client not configured; run register or login first")

const dirName = ".gophkeeper"

// Dir returns the client config directory in the user's home folder.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, dirName), nil
}

// EnsureDir creates the config directory if needed.
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// LoadConfig reads client configuration from disk.
func LoadConfig() (*config.ClientConfig, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &config.ClientConfig{}, nil
		}
		return nil, err
	}
	var cfg config.ClientConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// SaveConfig writes client configuration to disk.
func SaveConfig(cfg *config.ClientConfig) error {
	dir, err := EnsureDir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.json"), data, 0o600)
}

// LocalCache stores decrypted secret metadata for sync.
type LocalCache struct {
	LastSync time.Time        `json:"last_sync"`
	Secrets  []*model.Secret  `json:"secrets"`
}

// LoadCache reads the local sync cache.
func LoadCache() (*LocalCache, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, "cache.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &LocalCache{}, nil
		}
		return nil, err
	}
	var cache LocalCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

// SaveCache writes the local sync cache.
func SaveCache(cache *LocalCache) error {
	dir, err := EnsureDir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "cache.json"), data, 0o600)
}

// MergeSecrets merges remote changes into the local cache using last-write-wins by version.
func MergeSecrets(local []*model.Secret, remote []*model.Secret) []*model.Secret {
	byID := make(map[string]*model.Secret, len(local)+len(remote))
	for _, s := range local {
		cp := *s
		byID[s.ID] = &cp
	}
	for _, s := range remote {
		existing, ok := byID[s.ID]
		if !ok || s.Version >= existing.Version {
			cp := *s
			byID[s.ID] = &cp
		}
	}
	out := make([]*model.Secret, 0, len(byID))
	for _, s := range byID {
		if !s.IsDeleted() {
			out = append(out, s)
		}
	}
	return out
}

// FormatSecretList returns a human-readable list of secrets.
func FormatSecretList(secrets []*model.Secret) string {
	if len(secrets) == 0 {
		return "No secrets found."
	}
	var b string
	for _, s := range secrets {
		b += fmt.Sprintf("%s  %s  v%d  %s\n", s.ID, s.Name, s.Version, s.UpdatedAt.Format(time.RFC3339))
	}
	return b
}
