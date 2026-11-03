package model

import (
	"github.com/google/uuid"
)

const (
	// AdministratorAuthorityID 是固定管理员角色 ID。
	AdministratorAuthorityID uint = 888
	// UserAuthorityID 是普通用户角色 ID。
	UserAuthorityID uint = 9528
)

// User 表示系统登录用户及其固定角色边界。
type User struct {
	Timestamps            // Timestamps 保存用户记录的生命周期字段。
	UUID        uuid.UUID `json:"uuid" gorm:"index;comment:用户UUID"`                      // UUID 是跨接口暴露的用户稳定标识。
	Username    string    `json:"userName" gorm:"index;comment:用户登录名"`                   // Username 是唯一登录名。
	Password    string    `json:"-" gorm:"comment:用户登录密码"`                               // Password 是 bcrypt 密码摘要，不向客户端返回。
	NickName    string    `json:"nickName" gorm:"default:系统用户;comment:用户昵称"`             // NickName 是用户展示名称。
	HeaderImage string    `json:"headerImg" gorm:"column:header_img;comment:用户头像"`       // HeaderImage 是用户头像地址。
	AuthorityID uint      `json:"authorityId" gorm:"default:9528;comment:888管理员，其他普通用户"` // AuthorityID 是固定管理员或普通用户角色 ID。
	Phone       string    `json:"phone" gorm:"comment:用户手机号"`                            // Phone 是用户手机号。
	Email       string    `json:"email" gorm:"comment:用户邮箱"`                             // Email 是用户邮箱。
	Enable      int       `json:"enable" gorm:"default:1;comment:1正常 2冻结"`               // Enable 表示用户状态：1 正常，2 冻结。
}

// TableName 返回系统用户表名。
func (User) TableName() string {
	return "sys_users"
}
