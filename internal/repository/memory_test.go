package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/Irongoshan-ux/gophkeeper/internal/repository"
	"github.com/stretchr/testify/require"
)

func TestMemoryRepositoriesFullFlow(t *testing.T) {
	t.Parallel()
	users, secrets := repository.NewMemoryRepositories()

	_, err := users.Create(context.Background(), "dup", "h1")
	require.NoError(t, err)
	_, err = users.Create(context.Background(), "dup", "h2")
	require.ErrorIs(t, err, repository.ErrConflict)

	u, err := users.GetByID(context.Background(), "missing")
	require.ErrorIs(t, err, repository.ErrNotFound)
	require.Nil(t, u)

	owner, err := users.Create(context.Background(), "owner", "hash")
	require.NoError(t, err)

	created, err := secrets.Create(context.Background(), &model.Secret{
		UserID: owner.ID, Type: model.SecretTypeText, Name: "n", EncryptedData: []byte("x"), Version: 1,
	})
	require.NoError(t, err)

	list, err := secrets.List(context.Background(), owner.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	created.Name = "updated"
	created.Version = 1
	updated, err := secrets.Update(context.Background(), created)
	require.NoError(t, err)
	require.Equal(t, "updated", updated.Name)
	require.Equal(t, int64(2), updated.Version)

	now := time.Now()
	deleted, err := secrets.SoftDelete(context.Background(), owner.ID, created.ID, now, updated.Version)
	require.NoError(t, err)
	require.True(t, deleted.IsDeleted())

	since := time.Now().Add(-time.Hour)
	changed, err := secrets.ListSince(context.Background(), owner.ID, since)
	require.NoError(t, err)
	require.NotEmpty(t, changed)

	_, err = secrets.Get(context.Background(), owner.ID, created.ID)
	require.ErrorIs(t, err, repository.ErrNotFound)
}
