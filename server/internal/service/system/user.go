package system

import (
	"errors"
	"fmt"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model/system"
	systemReq "github.com/gcc798/xianyu.bot/internal/model/system/request"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct{ db *gorm.DB }

func (userService *UserService) Register(user system.User) (system.User, error) {
	var existing system.User
	err := userService.db.Where("username = ?", user.Username).First(&existing).Error
	if err == nil {
		return system.User{}, errors.New("用户名已注册")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return system.User{}, err
	}
	user.AuthorityID = normalizeAuthorityID(user.AuthorityID)
	if user.Enable == 0 {
		user.Enable = 1
	}
	user.Password = security.BcryptHash(user.Password)
	user.UUID = uuid.New()
	return user, userService.db.Create(&user).Error
}

func (userService *UserService) Login(input *system.User) (*system.User, error) {
	if userService.db == nil {
		return nil, fmt.Errorf("db not init")
	}
	var user system.User
	if err := userService.db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		return nil, err
	}
	if !security.BcryptCheck(input.Password, user.Password) {
		return nil, errors.New("密码错误")
	}
	return &user, nil
}

func (userService *UserService) ChangePassword(user *system.User, newPassword string) error {
	var stored system.User
	if err := userService.db.Select("id, password").Where("id = ?", user.ID).First(&stored).Error; err != nil {
		return err
	}
	if !security.BcryptCheck(user.Password, stored.Password) {
		return errors.New("原密码错误")
	}
	return userService.db.Model(&stored).Update("password", security.BcryptHash(newPassword)).Error
}

func (userService *UserService) GetUserInfoList(info systemReq.GetUserList) (list any, total int64, err error) {
	db := userService.db.Model(&system.User{})
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
	var users []system.User
	err = db.Limit(info.PageSize).Offset(info.PageSize * (info.Page - 1)).Order(order).Find(&users).Error
	return users, total, err
}

func (userService *UserService) DeleteUser(id int) error {
	return userService.db.Delete(&system.User{}, id).Error
}

func (userService *UserService) SetUserInfo(user system.User) error {
	user.AuthorityID = normalizeAuthorityID(user.AuthorityID)
	return userService.db.Model(&system.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"updated_at": time.Now(), "nick_name": user.NickName, "header_img": user.HeaderImage,
		"phone": user.Phone, "email": user.Email, "enable": user.Enable, "authority_id": user.AuthorityID,
	}).Error
}

func normalizeAuthorityID(authorityID uint) uint {
	if authorityID == system.AdministratorAuthorityID {
		return authorityID
	}
	return system.UserAuthorityID
}

func (userService *UserService) SetSelfInfo(user system.User) error {
	return userService.db.Model(&system.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"updated_at": time.Now(), "nick_name": user.NickName, "header_img": user.HeaderImage,
		"phone": user.Phone, "email": user.Email,
	}).Error
}

func (userService *UserService) GetUserInfo(userUUID uuid.UUID) (user system.User, err error) {
	err = userService.db.First(&user, "uuid = ?", userUUID).Error
	return user, err
}

func (userService *UserService) ResetPassword(id uint, password string) error {
	return userService.db.Model(&system.User{}).Where("id = ?", id).Update("password", security.BcryptHash(password)).Error
}
