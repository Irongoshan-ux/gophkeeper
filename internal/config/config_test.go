package config_test

import (
	"testing"

	"github.com/Irongoshan-ux/gophkeeper/internal/config"
	"github.com/stretchr/testify/require"
)

func TestDefaultServer(t *testing.T) {
	t.Parallel()
	cfg := config.DefaultServer()
	require.Equal(t, "localhost:9090", cfg.Address)
	require.NotEmpty(t, cfg.JWTSecret)
}
