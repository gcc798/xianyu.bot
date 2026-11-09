package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

// TestRevokeToken 验证对应测试场景。
func TestRevokeToken(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte("test"))
	require.NoError(t, err)
	require.False(t, IsTokenRevoked(token))
	RevokeToken(token)
	require.True(t, IsTokenRevoked(token))
}

// TestExpiredTokenIsNotRetained 验证对应测试场景。
func TestExpiredTokenIsNotRetained(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}).SignedString([]byte("test"))
	require.NoError(t, err)

	RevokeToken(token)
	require.False(t, IsTokenRevoked(token))
}
