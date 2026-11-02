package security

import (
	"testing"
	"time"

	"github.com/gcc798/go.xianyu.assistant/server/internal/config"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/system/request"
	"github.com/stretchr/testify/require"
)

func TestJWTUsesValidatedConfiguration(t *testing.T) {
	_, err := NewJWT(config.JWT{ExpiresTime: "invalid", BufferTime: "1h"})
	require.ErrorContains(t, err, "expires-time")

	manager, err := NewJWT(config.JWT{
		SigningKey: "test-secret", ExpiresTime: "2h", BufferTime: "15m", Issuer: "xianyu-test",
	})
	require.NoError(t, err)
	token, err := manager.CreateToken(manager.CreateClaims(request.BaseClaims{ID: 11, Username: "tester"}))
	require.NoError(t, err)
	claims, err := manager.ParseToken(token)
	require.NoError(t, err)
	require.Equal(t, uint(11), claims.BaseClaims.ID)
	require.Equal(t, "xianyu-test", claims.Issuer)
	require.EqualValues(t, (15 * time.Minute).Seconds(), claims.BufferTime)
	require.WithinDuration(t, time.Now().Add(2*time.Hour), claims.ExpiresAt.Time, time.Second)
}
