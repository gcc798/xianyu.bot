package service

import (
	"testing"

	model "github.com/gcc798/xianyu.bot/internal/model"
	"github.com/stretchr/testify/require"
)

// TestNormalizeAuthorityID 验证对应测试场景。
func TestNormalizeAuthorityID(t *testing.T) {
	require.Equal(t, model.AdministratorAuthorityID, normalizeAuthorityID(model.AdministratorAuthorityID))
	require.Equal(t, model.UserAuthorityID, normalizeAuthorityID(0))
	require.Equal(t, model.UserAuthorityID, normalizeAuthorityID(1234))
}
