package system

import (
	"testing"

	model "github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/stretchr/testify/require"
)

func TestNormalizeAuthorityID(t *testing.T) {
	require.Equal(t, model.AdministratorAuthorityID, normalizeAuthorityID(model.AdministratorAuthorityID))
	require.Equal(t, model.UserAuthorityID, normalizeAuthorityID(0))
	require.Equal(t, model.UserAuthorityID, normalizeAuthorityID(1234))
}
