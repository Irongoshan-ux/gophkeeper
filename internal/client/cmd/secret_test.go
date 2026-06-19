package cmd_test

import (
	"testing"

	"github.com/Irongoshan-ux/gophkeeper/internal/client/cmd"
	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/stretchr/testify/require"
)

func TestBuildPayloadCredentials(t *testing.T) {
	t.Parallel()
	st, payload, err := cmd.BuildPayload("credentials", "user", "pass", "", "", "")
	require.NoError(t, err)
	require.Equal(t, model.SecretTypeCredentials, st)
	require.Equal(t, "user", payload.Fields["login"])
}

func TestBuildPayloadCardInvalid(t *testing.T) {
	t.Parallel()
	_, _, err := cmd.BuildPayload("card", "", "", "", "1234", "")
	require.Error(t, err)
}
