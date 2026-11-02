package system

import (
	"errors"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct{}

var UserServiceApp = new(UserService)

func (userService *UserService) Register(user system.SysUser) (system.SysUser, error) {
	var existing system.SysUser
	err := global.GVA_DB.Where("username = ?", user.Username).First(&existing).Error
	if err == nil {
		return system.SysUser{}, errors.New("用户名已注册")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return system.SysUser{}, err
	}
	user.AuthorityId = normalizeAuthorityID(user.AuthorityId)
	if user.Enable == 0 {
		user.Enable = 1
	}
	user.Password = utils.BcryptHash(user.Password)
	user.UUID = uuid.New()
	return user, global.GVA_DB.Create(&user).Error
}

func (userService *UserService) Login(input *system.SysUser) (*system.SysUser, error) {
	if global.GVA_DB == nil {
		return nil, fmt.Errorf("db not init")
	}
	var user system.SysUser
	if err := global.GVA_DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		return nil, err
	}
	if !utils.BcryptCheck(input.Password, user.Password) {
		return nil, errors.New("密码错误")
	}
	return &user, nil
}

func (userService *UserService) ChangePassword(user *system.SysUser, newPassword string) error {
	var stored system.SysUser
	if err := global.GVA_DB.Select("id, password").Where("id = ?", user.ID).First(&stored).Error; err != nil {
		return err
	}
	if !utils.BcryptCheck(user.Password, stored.Password) {
		return errors.New("原密码错误")
	}
	return global.GVA_DB.Model(&stored).Update("password", utils.BcryptHash(newPassword)).Error
}

func (userService *UserService) GetUserInfoList(info systemReq.GetUserList) (list any, total int64, err error) {
	db := global.GVA_DB.Model(&system.SysUser{})
	for column, value := range map[string]string{
		"username":  info.Username,
		"nick_name": info.NickName,
		"phone":     info.Phone,
		"email":     info.Email,
	} {
		if value != "" {
			db = db.Where(column+" LIKE ?", "%"+value+"%")
		}
	}
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "id desc"
	allowedOrders := map[string]bool{"id": true, "username": true, "nick_name": true, "phone": true, "email": true}
	if allowedOrders[info.OrderKey] {
		order = info.OrderKey
		if info.Desc {
			order += " desc"
		}
	}
	var users []system.SysUser
	err = db.Limit(info.PageSize).Offset(info.PageSize * (info.Page - 1)).Order(order).Find(&users).Error
	return users, total, err
}

func (userService *UserService) DeleteUser(id int) error {
	return global.GVA_DB.Delete(&system.SysUser{}, id).Error
}

func (userService *UserService) SetUserInfo(user system.SysUser) error {
	user.AuthorityId = normalizeAuthorityID(user.AuthorityId)
	return global.GVA_DB.Model(&system.SysUser{}).Where("id = ?", user.ID).Updates(map[string]any{
		"updated_at": time.Now(), "nick_name": user.NickName, "header_img": user.HeaderImg,
		"phone": user.Phone, "email": user.Email, "enable": user.Enable, "authority_id": user.AuthorityId,
	}).Error
}

func normalizeAuthorityID(authorityID uint) uint {
	if authorityID == system.AdministratorAuthorityID {
		return authorityID
	}
	return system.UserAuthorityID
}

func (userService *UserService) SetSelfInfo(user system.SysUser) error {
	return global.GVA_DB.Model(&system.SysUser{}).Where("id = ?", user.ID).Updates(map[string]any{
		"updated_at": time.Now(), "nick_name": user.NickName, "header_img": user.HeaderImg,
		"phone": user.Phone, "email": user.Email,
	}).Error
}

func (userService *UserService) GetUserInfo(userUUID uuid.UUID) (user system.SysUser, err error) {
	err = global.GVA_DB.First(&user, "uuid = ?", userUUID).Error
	return user, err
}

func (userService *UserService) FindUserById(id int) (user *system.SysUser, err error) {
	user = new(system.SysUser)
	err = global.GVA_DB.Where("id = ?", id).First(user).Error
	return user, err
}

func (userService *UserService) FindUserByUuid(userUUID string) (user *system.SysUser, err error) {
	user = new(system.SysUser)
	if err = global.GVA_DB.Where("uuid = ?", userUUID).First(user).Error; err != nil {
		return user, errors.New("用户不存在")
	}
	return user, nil
}

func (userService *UserService) ResetPassword(id uint, password string) error {
	return global.GVA_DB.Model(&system.SysUser{}).Where("id = ?", id).Update("password", utils.BcryptHash(password)).Error
}
