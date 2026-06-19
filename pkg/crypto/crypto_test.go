package crypto_test

import (
	"testing"

	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/Irongoshan-ux/gophkeeper/pkg/crypto"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	t.Parallel()
	salt, err := crypto.NewSalt()
	require.NoError(t, err)
	key, err := crypto.DeriveKey("master", salt)
	require.NoError(t, err)

	payload := &model.Payload{
		Metadata: "github",
		Fields:   map[string]string{"login": "user", "password": "pass"},
	}
	encrypted, err := crypto.Encrypt(key, payload)
	require.NoError(t, err)

	decrypted, err := crypto.Decrypt(key, encrypted)
	require.NoError(t, err)
	require.Equal(t, payload.Metadata, decrypted.Metadata)
	require.Equal(t, payload.Fields["login"], decrypted.Fields["login"])
}

func TestDecryptWrongKey(t *testing.T) {
	t.Parallel()
	salt, err := crypto.NewSalt()
	require.NoError(t, err)
	key1, err := crypto.DeriveKey("one", salt)
	require.NoError(t, err)
	key2, err := crypto.DeriveKey("two", salt)
	require.NoError(t, err)

	encrypted, err := crypto.Encrypt(key1, &model.Payload{Fields: map[string]string{"x": "y"}})
	require.NoError(t, err)

	_, err = crypto.Decrypt(key2, encrypted)
	require.ErrorIs(t, err, crypto.ErrInvalidCiphertext)
}

func TestDeriveKeyEmptyPassword(t *testing.T) {
	t.Parallel()
	_, err := crypto.DeriveKey("", []byte("salt"))
	require.ErrorIs(t, err, crypto.ErrEmptyMasterPassword)
}
