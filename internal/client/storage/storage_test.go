package storage_test

import (
	"testing"
	"time"

	"github.com/Irongoshan-ux/gophkeeper/internal/client/storage"
	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/stretchr/testify/require"
)

func TestMergeSecretsLastWriteWins(t *testing.T) {
	t.Parallel()
	local := []*model.Secret{{
		ID: "1", Name: "old", Version: 1, UpdatedAt: time.Now().Add(-time.Hour),
	}}
	remote := []*model.Secret{{
		ID: "1", Name: "new", Version: 2, UpdatedAt: time.Now(),
	}}
	merged := storage.MergeSecrets(local, remote)
	require.Len(t, merged, 1)
	require.Equal(t, "new", merged[0].Name)
}

func TestMergeSecretsSkipsDeleted(t *testing.T) {
	t.Parallel()
	now := time.Now()
	deleted := now
	local := []*model.Secret{{ID: "1", Name: "gone", Version: 2, DeletedAt: &deleted, UpdatedAt: now}}
	merged := storage.MergeSecrets(local, nil)
	require.Len(t, merged, 0)
}
