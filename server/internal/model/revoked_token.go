package model

type RevokedToken struct {
	Timestamps
	Token string `gorm:"column:jwt;type:text;comment:jwt"`
}

func (RevokedToken) TableName() string { return "jwt_blacklists" }
