package response

import (
	"github.com/gcc798/xianyu.bot/internal/model"
)

// UserResponse 包装当前系统用户信息。
type UserResponse struct {
	User model.User `json:"user"` // User 是已认证用户资料。
}

// LoginResponse 返回登录用户、令牌及到期时间。
type LoginResponse struct {
	User      model.User `json:"user"`      // User 是已认证用户资料。
	Token     string     `json:"token"`     // Token 是本次签发的 JWT。
	ExpiresAt int64      `json:"expiresAt"` // ExpiresAt 是令牌到期的 Unix 时间戳。
}
