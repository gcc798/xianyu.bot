package security

import (
	"net"

	systemReq "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ClearToken(c *gin.Context) {
	// 增加cookie x-token 向来源的web添加
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

func SetToken(c *gin.Context, token string, maxAge int) {
	// 增加cookie x-token 向来源的web添加
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

func GetToken(c *gin.Context) string {
	token := c.Request.Header.Get("x-token")
	if token == "" {
		token, _ = c.Cookie("x-token")
	}
	return token
}

func claims(c *gin.Context) *systemReq.CustomClaims {
	value, exists := c.Get("claims")
	if !exists {
		return nil
	}
	claims, _ := value.(*systemReq.CustomClaims)
	return claims
}

// GetUserID 从Gin的Context中获取从jwt解析出来的用户ID
func GetUserID(c *gin.Context) uint {
	if claims := claims(c); claims != nil {
		return claims.BaseClaims.ID
	}
	return 0
}

// GetUserUuid 从Gin的Context中获取从jwt解析出来的用户UUID
func GetUserUuid(c *gin.Context) uuid.UUID {
	if claims := claims(c); claims != nil {
		return claims.UUID
	}
	return uuid.UUID{}
}

// GetUserAuthorityID 从Gin的Context中获取从jwt解析出来的用户角色id
func GetUserAuthorityID(c *gin.Context) uint {
	if claims := claims(c); claims != nil {
		return claims.AuthorityID
	}
	return 0
}
