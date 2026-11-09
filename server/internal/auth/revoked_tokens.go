package auth

import (
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// revokedTokens 在进程内保存令牌及其到期时间。
var revokedTokens sync.Map

// RevokeToken 将可解析且带到期时间的令牌加入当前进程的撤销集合。
func RevokeToken(token string) {
	claims := new(jwt.RegisteredClaims)
	if _, _, err := jwt.NewParser().ParseUnverified(token, claims); err != nil || claims.ExpiresAt == nil {
		return
	}
	revokedTokens.Store(token, claims.ExpiresAt.Time)
}

// IsTokenRevoked 判断令牌是否仍处于撤销有效期，并清理已到期记录。
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
