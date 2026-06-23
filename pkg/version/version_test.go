package version_test

import (
	"strings"
	"testing"

	"github.com/Irongoshan-ux/gophkeeper/pkg/version"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
	t.Parallel()
	info := version.Info()
	require.True(t, strings.Contains(info, "Build version:"))
	require.True(t, strings.Contains(info, "Build date:"))
}
