package system

import (
	"testing"

	model "github.com/gcc798/go.xianyu.assistant/server/internal/model/system"
	"github.com/stretchr/testify/require"
)

func TestNormalizeAuthorityID(t *testing.T) {
	require.Equal(t, model.AdministratorAuthorityID, normalizeAuthorityID(model.AdministratorAuthorityID))
	require.Equal(t, model.UserAuthorityID, normalizeAuthorityID(0))
	require.Equal(t, model.UserAuthorityID, normalizeAuthorityID(1234))
}
