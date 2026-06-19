package model_test

import (
	"testing"
	"time"

	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/stretchr/testify/require"
)

func TestSecretIsDeleted(t *testing.T) {
	t.Parallel()
	s := &model.Secret{}
	require.False(t, s.IsDeleted())
	now := time.Now()
	s.DeletedAt = &now
	require.True(t, s.IsDeleted())
}
