package system

import (
	"github.com/gcc798/xianyu.bot/internal/model/common"
	"github.com/google/uuid"
)

const (
	AdministratorAuthorityID uint = 888
	UserAuthorityID          uint = 9528
)

type SysUser struct {
	common.Timestamps
	UUID        uuid.UUID `json:"uuid" gorm:"index;comment:用户UUID"`
	Username    string    `json:"userName" gorm:"index;comment:用户登录名"`
	Password    string    `json:"-" gorm:"comment:用户登录密码"`
	NickName    string    `json:"nickName" gorm:"default:系统用户;comment:用户昵称"`
	HeaderImg   string    `json:"headerImg" gorm:"comment:用户头像"`
	AuthorityId uint      `json:"authorityId" gorm:"default:9528;comment:888管理员，其他普通用户"`
	Phone       string    `json:"phone" gorm:"comment:用户手机号"`
	Email       string    `json:"email" gorm:"comment:用户邮箱"`
	Enable      int       `json:"enable" gorm:"default:1;comment:1正常 2冻结"`
}

func (SysUser) TableName() string {
	return "sys_users"
}
