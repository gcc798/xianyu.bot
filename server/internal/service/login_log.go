package service

import (
	"github.com/gcc798/xianyu.bot/internal/model"
	systemReq "github.com/gcc798/xianyu.bot/internal/model/request"
	"gorm.io/gorm"
)

// LoginLogService 管理系统登录审计记录。
type LoginLogService struct {
	db *gorm.DB // db 持久化登录日志。
}

// NewLoginLogService 创建登录审计服务。
func NewLoginLogService(db *gorm.DB) *LoginLogService {
	return &LoginLogService{db: db}
}

// Create 保存一次登录尝试结果。
func (loginLogService *LoginLogService) Create(loginLog model.LoginLog) error {
	return loginLogService.db.Create(&loginLog).Error
}

// Delete 删除指定登录审计记录。
func (loginLogService *LoginLogService) Delete(loginLog model.LoginLog) error {
	return loginLogService.db.Delete(&loginLog).Error
}

// List 按 info 分页筛选登录审计并预加载关联用户。
func (loginLogService *LoginLogService) List(info systemReq.LoginLogSearch) ([]model.LoginLog, int64, error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := loginLogService.db.Model(&model.LoginLog{})
	var loginLogs []model.LoginLog
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
