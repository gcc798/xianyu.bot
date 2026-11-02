package response

import (
	"github.com/gcc798/xianyu.bot/internal/model/system"
)

type UserResponse struct {
	User system.User `json:"user"`
}

type LoginResponse struct {
	User      system.User `json:"user"`
	Token     string      `json:"token"`
	ExpiresAt int64       `json:"expiresAt"`
}
