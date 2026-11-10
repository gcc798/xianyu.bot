package middleware

import (
	"errors"
	"strconv"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gin-gonic/gin"
)

// JWTAuth 校验访问令牌并写入用户身份。
func JWTAuth(jwtManager *security.JWT) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 我们这里jwt鉴权取头部信息 x-token 登录时回返回token信息 这里前端需要把token存储到cookie或者本地localStorage中 不过需要跟后端协商过期时间 可以约定刷新令牌或者重新登录
		token := security.GetToken(c)
		if token == "" {
			response.NoAuth("未登录或非法访问，请登录", c)
			c.Abort()
			return
		}
		// parseToken 解析token包含的信息
		claims, err := jwtManager.ParseToken(token)
		if err != nil {
			if errors.Is(err, security.TokenExpired) {
				response.NoAuth("登录已过期，请重新登录", c)
				security.ClearToken(c)
				c.Abort()
				return
			}
			response.NoAuth(err.Error(), c)
			security.ClearToken(c)
			c.Abort()
			return
		}

		c.Set("claims", claims)
		if claims.ExpiresAt.Unix()-time.Now().Unix() < claims.BufferTime {
			dr := jwtManager.ExpiresIn()
			claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(dr))
			newToken, _ := jwtManager.CreateTokenByOldToken(token, *claims)
			newClaims, _ := jwtManager.ParseToken(newToken)
			c.Header("new-token", newToken)
			c.Header("new-expires-at", strconv.FormatInt(newClaims.ExpiresAt.Unix(), 10))
			security.SetToken(c, newToken, int(dr.Seconds()))
		}
		c.Next()

		if newToken, exists := c.Get("new-token"); exists {
			c.Header("new-token", newToken.(string))
		}
		if newExpiresAt, exists := c.Get("new-expires-at"); exists {
			c.Header("new-expires-at", newExpiresAt.(string))
		}
	}
}

// AdminOnly 限制请求仅管理员访问。
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if security.GetUserAuthorityID(c) != model.AdministratorAuthorityID {
			c.JSON(403, response.Response{Code: response.ERROR, Msg: "仅管理员可操作"})
			c.Abort()
			return
		}
		c.Next()
	}
}
