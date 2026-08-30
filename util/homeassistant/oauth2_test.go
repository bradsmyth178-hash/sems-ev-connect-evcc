package homeassistant

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSupervisorTokenSource(t *testing.T) {
	t.Setenv("SUPERVISOR_TOKEN", "supervisor-secret")

	source, ok := supervisorTokenSource("http://supervisor/core")
	require.True(t, ok)
	token, err := source.Token()
	require.NoError(t, err)
	require.Equal(t, "supervisor-secret", token.AccessToken)
	require.Equal(t, "Bearer", token.TokenType)
}

func TestSupervisorTokenIsRestrictedToInternalCoreProxy(t *testing.T) {
	t.Setenv("SUPERVISOR_TOKEN", "supervisor-secret")

	for _, uri := range []string{
		"https://supervisor/core",
		"http://supervisor/other",
		"http://supervisor.example/core",
		"http://example.com/core",
	} {
		t.Run(uri, func(t *testing.T) {
			_, ok := supervisorTokenSource(uri)
			require.False(t, ok)
		})
	}
}

func TestSupervisorTokenSourceRequiresToken(t *testing.T) {
	t.Setenv("SUPERVISOR_TOKEN", "")
	_, ok := supervisorTokenSource("http://supervisor/core")
	require.False(t, ok)
}
