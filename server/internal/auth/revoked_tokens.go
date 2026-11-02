package auth

import (
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var revokedTokens sync.Map

func RevokeToken(token string) {
	claims := new(jwt.RegisteredClaims)
	if _, _, err := jwt.NewParser().ParseUnverified(token, claims); err != nil || claims.ExpiresAt == nil {
		return
	}
	revokedTokens.Store(token, claims.ExpiresAt.Time)
}

func IsTokenRevoked(token string) bool {
	expiresAt, revoked := revokedTokens.Load(token)
	if !revoked {
		return false
	}
	if time.Now().After(expiresAt.(time.Time)) {
		revokedTokens.Delete(token)
		return false
	}
	return true
}
