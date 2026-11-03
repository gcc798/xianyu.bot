package model

// RevokedToken 保存需要在持久化层失效的 JWT。
type RevokedToken struct {
	// Timestamps 保存撤销记录的生命周期字段。
	Timestamps
	// Token 是已撤销的完整 JWT。
	Token string `gorm:"column:jwt;type:text;comment:jwt"`
}

// TableName 返回兼容旧系统的令牌黑名单表名。
func (RevokedToken) TableName() string { return "jwt_blacklists" }
