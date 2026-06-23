package validation_test

import (
	"testing"

	"github.com/Irongoshan-ux/gophkeeper/internal/validation"
	"github.com/stretchr/testify/require"
)

func TestCredentials(t *testing.T) {
	t.Parallel()
	require.ErrorIs(t, validation.Credentials("", "pass"), validation.ErrEmptyLogin)
	require.ErrorIs(t, validation.Credentials("user", ""), validation.ErrEmptyPassword)
	require.NoError(t, validation.Credentials("user", "pass"))
}

func TestLuhnValid(t *testing.T) {
	t.Parallel()
	require.NoError(t, validation.Luhn("4532015112830366"))
}

func TestLuhnInvalid(t *testing.T) {
	t.Parallel()
	require.ErrorIs(t, validation.Luhn("1234567890123456"), validation.ErrInvalidCard)
}

func TestSecretName(t *testing.T) {
	t.Parallel()
	require.ErrorIs(t, validation.SecretName("  "), validation.ErrEmptyName)
	require.NoError(t, validation.SecretName("my-secret"))
}
