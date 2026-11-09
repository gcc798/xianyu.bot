package security

import (
	"net"

	systemReq "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ClearToken 清除请求来源域下的 x-token Cookie。
func ClearToken(c *gin.Context) {
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}

	if net.ParseIP(host) != nil {
		c.SetCookie("x-token", "", -1, "/", "", false, false)
	} else {
		c.SetCookie("x-token", "", -1, "/", host, false, false)
	}
}

// SetToken 将 token 写入请求来源域，maxAge 以秒为单位。
func SetToken(c *gin.Context, token string, maxAge int) {
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}

	if net.ParseIP(host) != nil {
		c.SetCookie("x-token", token, maxAge, "/", "", false, false)
	} else {
		c.SetCookie("x-token", token, maxAge, "/", host, false, false)
	}
}

// GetToken 优先读取 x-token 请求头，缺失时回退到同名 Cookie。
func GetToken(c *gin.Context) string {
	token := c.Request.Header.Get("x-token")
	if token == "" {
		token, _ = c.Cookie("x-token")
	}
	return token
}

// claims 读取认证中间件写入 Gin 上下文的令牌声明。
func claims(c *gin.Context) *systemReq.CustomClaims {
	value, exists := c.Get("claims")
	if !exists {
		return nil
	}
	claims, _ := value.(*systemReq.CustomClaims)
	return claims
}

// GetUserID 返回认证用户的数据库主键，未认证时返回零。
func GetUserID(c *gin.Context) uint {
	if claims := claims(c); claims != nil {
		return claims.BaseClaims.ID
	}
	return 0
}

// GetUserUuid 返回认证用户的 UUID，未认证时返回零值。
func GetUserUuid(c *gin.Context) uuid.UUID {
	if claims := claims(c); claims != nil {
		return claims.UUID
	}
	return uuid.UUID{}
}

// GetUserAuthorityID 返回认证用户的角色 ID，未认证时返回零。
func GetUserAuthorityID(c *gin.Context) uint {
	if claims := claims(c); claims != nil {
		return claims.AuthorityID
	}
	return 0
}
