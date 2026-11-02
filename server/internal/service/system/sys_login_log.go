package system

import (
	"github.com/gcc798/xianyu.bot/internal/model/system"
	systemReq "github.com/gcc798/xianyu.bot/internal/model/system/request"
	"gorm.io/gorm"
)

type LoginLogService struct{ db *gorm.DB }

func (loginLogService *LoginLogService) CreateLoginLog(loginLog system.SysLoginLog) (err error) {
	err = loginLogService.db.Create(&loginLog).Error
	return err
}

func (loginLogService *LoginLogService) DeleteLoginLog(loginLog system.SysLoginLog) (err error) {
	err = loginLogService.db.Delete(&loginLog).Error
	return err
}

func (loginLogService *LoginLogService) GetLoginLogInfoList(info systemReq.SysLoginLogSearch) (list interface{}, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	// 创建db
	db := loginLogService.db.Model(&system.SysLoginLog{})
	var loginLogs []system.SysLoginLog
	// 如果有条件搜索 下方会自动创建搜索语句
	if info.Username != "" {
		db = db.Where("username LIKE ?", "%"+info.Username+"%")
	}
	if info.Status != false {
		db = db.Where("status = ?", info.Status)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Limit(limit).Offset(offset).Order("id desc").Preload("User").Find(&loginLogs).Error
	return loginLogs, total, err
}
