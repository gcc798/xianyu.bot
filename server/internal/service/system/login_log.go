package system

import (
	"github.com/gcc798/xianyu.bot/internal/model/system"
	systemReq "github.com/gcc798/xianyu.bot/internal/model/system/request"
	"gorm.io/gorm"
)

type LoginLogService struct{ db *gorm.DB }

func (loginLogService *LoginLogService) Create(loginLog system.LoginLog) error {
	return loginLogService.db.Create(&loginLog).Error
}

func (loginLogService *LoginLogService) Delete(loginLog system.LoginLog) error {
	return loginLogService.db.Delete(&loginLog).Error
}

func (loginLogService *LoginLogService) List(info systemReq.LoginLogSearch) ([]system.LoginLog, int64, error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := loginLogService.db.Model(&system.LoginLog{})
	var loginLogs []system.LoginLog
	if info.Username != "" {
		db = db.Where("username LIKE ?", "%"+info.Username+"%")
	}
	if info.Status != false {
		db = db.Where("status = ?", info.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Limit(limit).Offset(offset).Order("id desc").Preload("User").Find(&loginLogs).Error
	return loginLogs, total, err
}
