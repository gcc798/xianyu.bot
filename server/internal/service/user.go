package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	systemReq "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct{ db *gorm.DB }

func (userService *UserService) Register(user model.User) (model.User, error) {
	var existing model.User
	err := userService.db.Where("username = ?", user.Username).First(&existing).Error
	if err == nil {
		return model.User{}, errors.New("用户名已注册")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, err
	}
	user.AuthorityID = normalizeAuthorityID(user.AuthorityID)
	if user.Enable == 0 {
		user.Enable = 1
	}
	user.Password = security.BcryptHash(user.Password)
	user.UUID = uuid.New()
	return user, userService.db.Create(&user).Error
}

func (userService *UserService) Login(input *model.User) (*model.User, error) {
	if userService.db == nil {
		return nil, fmt.Errorf("db not init")
	}
	var user model.User
	if err := userService.db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		return nil, err
	}
	if !security.BcryptCheck(input.Password, user.Password) {
		return nil, errors.New("密码错误")
	}
	return &user, nil
}

func (userService *UserService) ChangePassword(user *model.User, newPassword string) error {
	var stored model.User
	if err := userService.db.Select("id, password").Where("id = ?", user.ID).First(&stored).Error; err != nil {
		return err
	}
	if !security.BcryptCheck(user.Password, stored.Password) {
		return errors.New("原密码错误")
	}
	return userService.db.Model(&stored).Update("password", security.BcryptHash(newPassword)).Error
}

func (userService *UserService) List(info systemReq.UserSearch) ([]model.User, int64, error) {
	db := userService.db.Model(&model.User{})
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
	var total int64
	if err := db.Count(&total).Error; err != nil {
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
	var users []model.User
	if err := db.Scopes(info.Paginate()).Order(order).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (userService *UserService) DeleteUser(id int) error {
	return userService.db.Delete(&model.User{}, id).Error
}

func (userService *UserService) SetUserInfo(user model.User) error {
	user.AuthorityID = normalizeAuthorityID(user.AuthorityID)
	return userService.db.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"updated_at": time.Now(), "nick_name": user.NickName, "header_img": user.HeaderImage,
		"phone": user.Phone, "email": user.Email, "enable": user.Enable, "authority_id": user.AuthorityID,
	}).Error
}

func normalizeAuthorityID(authorityID uint) uint {
	if authorityID == model.AdministratorAuthorityID {
		return authorityID
	}
	return model.UserAuthorityID
}

func (userService *UserService) SetSelfInfo(user model.User) error {
	return userService.db.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"updated_at": time.Now(), "nick_name": user.NickName, "header_img": user.HeaderImage,
		"phone": user.Phone, "email": user.Email,
	}).Error
}

func (userService *UserService) GetUserInfo(userUUID uuid.UUID) (user model.User, err error) {
	err = userService.db.First(&user, "uuid = ?", userUUID).Error
	return user, err
}

func (userService *UserService) ResetPassword(id uint, password string) error {
	return userService.db.Model(&model.User{}).Where("id = ?", id).Update("password", security.BcryptHash(password)).Error
}
