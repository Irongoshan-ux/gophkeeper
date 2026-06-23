package grpcserver_test

import (
	"context"
	"net"
	"testing"

	gophkeeperv1 "github.com/Irongoshan-ux/gophkeeper/api/gophkeeper/v1"
	"github.com/Irongoshan-ux/gophkeeper/internal/grpcserver"
	"github.com/Irongoshan-ux/gophkeeper/internal/repository"
	"github.com/Irongoshan-ux/gophkeeper/internal/service"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupTestServer(t *testing.T) (gophkeeperv1.AuthServiceClient, gophkeeperv1.SecretServiceClient, func()) {
	t.Helper()
	users, secrets := repository.NewMemoryRepositories()
	authSvc := service.NewAuthService(users, "test-secret")
	secretSvc := service.NewSecretService(secrets)
	handler := grpcserver.NewServer(authSvc, secretSvc, zerolog.Nop())

	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.AuthUnaryInterceptor("test-secret")),
	)
	gophkeeperv1.RegisterAuthServiceServer(srv, handler)
	gophkeeperv1.RegisterSecretServiceServer(srv, handler)

	go func() { _ = srv.Serve(lis) }()

	dialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		srv.Stop()
		_ = conn.Close()
	}
	return gophkeeperv1.NewAuthServiceClient(conn), gophkeeperv1.NewSecretServiceClient(conn), cleanup
}

func authCtx(token string) context.Context {
	return metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+token)
}

func TestRegisterLoginAndSecretFlow(t *testing.T) {
	t.Parallel()
	authClient, secretClient, cleanup := setupTestServer(t)
	defer cleanup()

	reg, err := authClient.Register(context.Background(), &gophkeeperv1.RegisterRequest{
		Login: "alice", Password: "pass",
	})
	require.NoError(t, err)
	require.NotEmpty(t, reg.GetToken())

	login, err := authClient.Login(context.Background(), &gophkeeperv1.LoginRequest{
		Login: "alice", Password: "pass",
	})
	require.NoError(t, err)
	require.NotEmpty(t, login.GetToken())

	ctx := authCtx(reg.GetToken())
	created, err := secretClient.CreateSecret(ctx, &gophkeeperv1.CreateSecretRequest{
		Type: gophkeeperv1.SecretType_SECRET_TYPE_TEXT, Name: "note", EncryptedData: []byte("cipher"),
	})
	require.NoError(t, err)
	require.NotEmpty(t, created.GetId())

	got, err := secretClient.GetSecret(ctx, &gophkeeperv1.GetSecretRequest{Id: created.GetId()})
	require.NoError(t, err)
	require.Equal(t, "note", got.GetName())

	list, err := secretClient.ListSecrets(ctx, &gophkeeperv1.ListSecretsRequest{})
	require.NoError(t, err)
	require.Len(t, list.GetSecrets(), 1)

	updated, err := secretClient.UpdateSecret(ctx, &gophkeeperv1.UpdateSecretRequest{
		Id: created.GetId(), Type: gophkeeperv1.SecretType_SECRET_TYPE_TEXT,
		Name: "note2", EncryptedData: []byte("cipher2"), Version: created.GetVersion(),
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), updated.GetVersion())

	_, err = secretClient.DeleteSecret(ctx, &gophkeeperv1.DeleteSecretRequest{
		Id: created.GetId(), Version: updated.GetVersion(),
	})
	require.NoError(t, err)
}

func TestUnauthenticatedSecretRejected(t *testing.T) {
	t.Parallel()
	_, secretClient, cleanup := setupTestServer(t)
	defer cleanup()

	_, err := secretClient.ListSecrets(context.Background(), &gophkeeperv1.ListSecretsRequest{})
	require.Error(t, err)
}

func TestRegisterDuplicate(t *testing.T) {
	t.Parallel()
	authClient, _, cleanup := setupTestServer(t)
	defer cleanup()

	_, err := authClient.Register(context.Background(), &gophkeeperv1.RegisterRequest{
		Login: "dup", Password: "pass",
	})
	require.NoError(t, err)

	_, err = authClient.Register(context.Background(), &gophkeeperv1.RegisterRequest{
		Login: "dup", Password: "pass",
	})
	require.Error(t, err)
	require.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestLoginInvalidCredentials(t *testing.T) {
	t.Parallel()
	authClient, _, cleanup := setupTestServer(t)
	defer cleanup()

	_, err := authClient.Register(context.Background(), &gophkeeperv1.RegisterRequest{
		Login: "user", Password: "pass",
	})
	require.NoError(t, err)

	_, err = authClient.Login(context.Background(), &gophkeeperv1.LoginRequest{
		Login: "user", Password: "wrong",
	})
	require.Error(t, err)
	require.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestSyncRPC(t *testing.T) {
	t.Parallel()
	authClient, secretClient, cleanup := setupTestServer(t)
	defer cleanup()

	reg, err := authClient.Register(context.Background(), &gophkeeperv1.RegisterRequest{
		Login: "sync", Password: "pass",
	})
	require.NoError(t, err)

	ctx := authCtx(reg.GetToken())
	_, err = secretClient.CreateSecret(ctx, &gophkeeperv1.CreateSecretRequest{
		Type: gophkeeperv1.SecretType_SECRET_TYPE_TEXT, Name: "s", EncryptedData: []byte("x"),
	})
	require.NoError(t, err)

	resp, err := secretClient.Sync(ctx, &gophkeeperv1.SyncRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, resp.GetSecrets())
}
