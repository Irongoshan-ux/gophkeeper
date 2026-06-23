package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	t.Parallel()
	cert, err := generateSelfSignedCert()
	require.NoError(t, err)
	require.NotEmpty(t, cert.Certificate)
}
