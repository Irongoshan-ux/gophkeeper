package app_test

import (
	"testing"

	"github.com/Irongoshan-ux/gophkeeper/internal/app"
	"github.com/stretchr/testify/require"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	t.Parallel()
	cert, err := app.GenerateSelfSignedCert()
	require.NoError(t, err)
	require.NotEmpty(t, cert.Certificate)
}
