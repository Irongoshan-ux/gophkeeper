package client_test

import (
	"context"
	"net"
	"testing"
	"time"

	gophkeeperv1 "github.com/Irongoshan-ux/gophkeeper/api/gophkeeper/v1"
	"github.com/Irongoshan-ux/gophkeeper/internal/client"
	"github.com/Irongoshan-ux/gophkeeper/internal/grpcserver"
	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/Irongoshan-ux/gophkeeper/internal/repository"
	"github.com/Irongoshan-ux/gophkeeper/internal/service"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func setupAPI(t *testing.T) *client.API {
	t.Helper()
	users, secrets := repository.NewMemoryRepositories()
	authSvc := service.NewAuthService(users, "secret")
	secretSvc := service.NewSecretService(secrets)
	handler := grpcserver.NewServer(authSvc, secretSvc, zerolog.Nop())

	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer(grpc.UnaryInterceptor(grpcserver.AuthUnaryInterceptor("secret")))
	gophkeeperv1.RegisterAuthServiceServer(srv, handler)
	gophkeeperv1.RegisterSecretServiceServer(srv, handler)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	dialer := func(context.Context, string) (net.Conn, error) { return lis.Dial() }
	api, err := client.Connect(client.Options{
		Address:  "passthrough:///bufnet",
		Insecure: true,
		Dialer:   dialer,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = api.Close() })
	return api
}

func TestClientRegisterAndSecretCRUD(t *testing.T) {
	t.Parallel()
	api := setupAPI(t)

	token, userID, err := api.Register(context.Background(), "cli", "pass")
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, userID)

	secret, err := api.CreateSecret(context.Background(), model.SecretTypeText, "note", []byte("enc"))
	require.NoError(t, err)

	got, err := api.GetSecret(context.Background(), secret.ID)
	require.NoError(t, err)
	require.Equal(t, "note", got.Name)

	list, err := api.ListSecrets(context.Background())
	require.NoError(t, err)
	require.Len(t, list, 1)

	updated, err := api.UpdateSecret(context.Background(), secret.ID, model.SecretTypeText, "note2", []byte("enc2"), secret.Version)
	require.NoError(t, err)
	require.Equal(t, int64(2), updated.Version)

	require.NoError(t, api.DeleteSecret(context.Background(), secret.ID, updated.Version))

	synced, err := api.Sync(context.Background(), updated.UpdatedAt.Add(-time.Hour))
	require.NoError(t, err)
	require.NotEmpty(t, synced)
}
