package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Irongoshan-ux/gophkeeper/internal/auth"
	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/Irongoshan-ux/gophkeeper/internal/repository"
	"github.com/Irongoshan-ux/gophkeeper/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAuthRegisterLogin(t *testing.T) {
	t.Parallel()
	users, secrets := repository.NewMemoryRepositories()
	_ = secrets
	authSvc := service.NewAuthService(users, "test-secret")

	token, userID, err := authSvc.Register(context.Background(), "alice", "password")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, userID)

	_, _, err = authSvc.Register(context.Background(), "alice", "password")
	require.ErrorIs(t, err, repository.ErrConflict)

	token2, userID2, err := authSvc.Login(context.Background(), "alice", "password")
	require.NoError(t, err)
	require.NotEmpty(t, token2)
	require.Equal(t, userID, userID2)

	_, _, err = authSvc.Login(context.Background(), "alice", "wrong")
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestSecretCRUD(t *testing.T) {
	t.Parallel()
	users, secrets := repository.NewMemoryRepositories()
	authSvc := service.NewAuthService(users, "test-secret")
	secretSvc := service.NewSecretService(secrets)

	_, userID, err := authSvc.Register(context.Background(), "bob", "password")
	require.NoError(t, err)

	ctx := auth.WithUserID(context.Background(), userID)
	created, err := secretSvc.Create(ctx, model.SecretTypeText, "note", []byte("cipher"))
	require.NoError(t, err)
	require.Equal(t, int64(1), created.Version)

	got, err := secretSvc.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "note", got.Name)

	list, err := secretSvc.List(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)

	updated, err := secretSvc.Update(ctx, created.ID, model.SecretTypeText, "note2", []byte("cipher2"), created.Version)
	require.NoError(t, err)
	require.Equal(t, int64(2), updated.Version)

	deleted, err := secretSvc.Delete(ctx, created.ID, updated.Version)
	require.NoError(t, err)
	require.True(t, deleted.IsDeleted())

	_, err = secretSvc.Get(ctx, created.ID)
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestSecretSync(t *testing.T) {
	t.Parallel()
	users, secrets := repository.NewMemoryRepositories()
	authSvc := service.NewAuthService(users, "test-secret")
	secretSvc := service.NewSecretService(secrets)

	_, userID, err := authSvc.Register(context.Background(), "sync-user", "password")
	require.NoError(t, err)

	ctx := auth.WithUserID(context.Background(), userID)
	_, err = secretSvc.Create(ctx, model.SecretTypeText, "a", []byte("x"))
	require.NoError(t, err)

	all, err := secretSvc.Sync(ctx, time.Time{})
	require.NoError(t, err)
	require.Len(t, all, 1)
}

func TestSecretUpdateConflict(t *testing.T) {
	t.Parallel()
	_, secrets := repository.NewMemoryRepositories()
	secretSvc := service.NewSecretService(secrets)

	ctx := auth.WithUserID(context.Background(), "user-1")
	created, err := secretSvc.Create(ctx, model.SecretTypeText, "x", []byte("a"))
	require.NoError(t, err)

	_, err = secretSvc.Update(ctx, created.ID, model.SecretTypeText, "y", []byte("b"), 999)
	require.ErrorIs(t, err, repository.ErrConflict)
}
