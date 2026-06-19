package auth_test

import (
	"context"
	"testing"

	"github.com/Irongoshan-ux/gophkeeper/internal/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

func TestNewTokenAndParse(t *testing.T) {
	t.Parallel()
	token, err := auth.NewToken("secret", "user-1")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	userID, err := auth.ParseToken("secret", token)
	require.NoError(t, err)
	require.Equal(t, "user-1", userID)
}

func TestParseTokenInvalid(t *testing.T) {
	t.Parallel()
	_, err := auth.ParseToken("secret", "bad-token")
	require.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestUserIDFromMetadata(t *testing.T) {
	t.Parallel()
	token, err := auth.NewToken("secret", "user-42")
	require.NoError(t, err)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
	userID, err := auth.UserIDFromMetadata(ctx, "secret")
	require.NoError(t, err)
	require.Equal(t, "user-42", userID)
}

func TestUserIDFromContext(t *testing.T) {
	t.Parallel()
	ctx := auth.WithUserID(context.Background(), "abc")
	userID, err := auth.UserIDFromContext(ctx)
	require.NoError(t, err)
	require.Equal(t, "abc", userID)
}

func TestOutgoingContext(t *testing.T) {
	t.Parallel()
	ctx := auth.OutgoingContext(context.Background(), "tok")
	require.NotNil(t, ctx)
}
