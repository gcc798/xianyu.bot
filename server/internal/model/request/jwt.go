package request

import (
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// CustomClaims 包含业务身份、刷新缓冲期和标准 JWT 声明。
type CustomClaims struct {
	// BaseClaims 保存系统用户身份。
	BaseClaims
	// BufferTime 是允许刷新令牌的缓冲秒数。
	BufferTime int64
	// RegisteredClaims 保存标准 JWT 时间与签发方声明。
	jwt.RegisteredClaims
}

// BaseClaims 是服务鉴权所需的最小用户身份快照。
type BaseClaims struct {
	// UUID 是用户对外稳定标识。
	UUID uuid.UUID
	// ID 是用户数据库主键。
	ID uint
	// Username 是用户登录名。
	Username string
	// NickName 是用户显示名称。
	NickName string
	// AuthorityID 是用户固定角色 ID。
	AuthorityID uint `json:"AuthorityId"`
}
