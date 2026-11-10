package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"gorm.io/gorm"
)

// TelegramAPIConfig 管理 Telegram 应用凭据及限流状态。
type TelegramAPIConfig struct {
	db *gorm.DB
}

// NewTelegramAPIConfig 创建 Telegram API 配置服务。
func NewTelegramAPIConfig(db *gorm.DB) *TelegramAPIConfig {
	return &TelegramAPIConfig{db: db}
}

// List 返回匹配条件的 Telegram API 配置，不解密凭据。
func (s *TelegramAPIConfig) List(ctx context.Context, input xyRequest.TelegramAPIConfigSearch) ([]model.TelegramAPIConfig, int64, error) {
	db := s.db.WithContext(ctx).Model(&model.TelegramAPIConfig{})
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		db = db.Where("LOWER(name) LIKE LOWER(?) OR CAST(api_id AS TEXT) LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	configs := make([]model.TelegramAPIConfig, 0)
	err := db.Scopes(input.Paginate()).Order("api_id DESC").Find(&configs).Error
	for i := range configs {
		configs[i].HasAPIHash = configs[i].APIHash != ""
	}
	return configs, total, err
}

// Create 新增 Telegram API 配置并加密应用 Hash。
func (s *TelegramAPIConfig) Create(ctx context.Context, input xyRequest.SaveTelegramAPIConfig) (model.TelegramAPIConfig, error) {
	if err := validateTelegramAPIConfig(input, true); err != nil {
		return model.TelegramAPIConfig{}, err
	}
	config := model.TelegramAPIConfig{Name: strings.TrimSpace(input.Name), APIID: input.APIID}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&config).Error; err != nil {
			return err
		}
		ciphertext, err := encryptCredential(tx, strings.TrimSpace(input.APIHash), telegramAPIHashAAD(config.ID))
		if err != nil {
			return err
		}
		return tx.Model(&config).Update("api_hash", ciphertext).Error
	})
	config.HasAPIHash = err == nil
	return config, err
}

// Update 修改配置；空 Hash 保留原凭据及限流状态。
func (s *TelegramAPIConfig) Update(ctx context.Context, id uint, input xyRequest.SaveTelegramAPIConfig) error {
	if id == 0 {
		return errors.New("Telegram API 配置 ID 无效")
	}
	if err := validateTelegramAPIConfig(input, false); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.TelegramAPIConfig
		if err := tx.First(&current, id).Error; err != nil {
			return err
		}
		updates := map[string]any{"name": strings.TrimSpace(input.Name), "api_id": input.APIID}
		if strings.TrimSpace(input.APIHash) != "" {
			ciphertext, err := encryptCredential(tx, strings.TrimSpace(input.APIHash), telegramAPIHashAAD(id))
			if err != nil {
				return err
			}
			updates["api_hash"] = ciphertext
		}
		return tx.Model(&current).Updates(updates).Error
	})
}

// Delete 删除指定配置。
func (s *TelegramAPIConfig) Delete(ctx context.Context, id uint) error {
	if id == 0 {
		return errors.New("Telegram API 配置 ID 无效")
	}
	result := s.db.WithContext(ctx).Unscoped().Delete(&model.TelegramAPIConfig{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Active 返回 API ID 最大且当前未限流的配置。
func (s *TelegramAPIConfig) Active(ctx context.Context) (model.TelegramAPIConfig, error) {
	var config model.TelegramAPIConfig
	err := s.db.WithContext(ctx).
		Where("api_hash <> '' AND (flood_wait_until IS NULL OR flood_wait_until <= ?)", time.Now()).
		Order("api_id DESC").First(&config).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return model.TelegramAPIConfig{}, err
		}
		var earliest *time.Time
		if queryErr := s.db.WithContext(ctx).Model(&model.TelegramAPIConfig{}).
			Where("api_hash <> '' AND flood_wait_until IS NOT NULL").
			Select("MIN(flood_wait_until)").Scan(&earliest).Error; queryErr != nil {
			return model.TelegramAPIConfig{}, queryErr
		}
		if earliest != nil {
			return model.TelegramAPIConfig{}, fmt.Errorf("所有 Telegram API 配置均限流，最早可用时间 %s", earliest.Format(time.RFC3339))
		}
		return model.TelegramAPIConfig{}, errors.New("尚未添加 Telegram API 配置")
	}
	config.APIHash, err = decryptCredential(s.db.WithContext(ctx), config.APIHash, telegramAPIHashAAD(config.ID))
	if err != nil {
		return model.TelegramAPIConfig{}, fmt.Errorf("decrypt Telegram API Hash: %w", err)
	}
	return config, nil
}

// RecordFloodWait 保存 Telegram 返回的限流时间。
func (s *TelegramAPIConfig) RecordFloodWait(ctx context.Context, apiID, seconds int64) error {
	if apiID <= 0 || seconds <= 0 || seconds > int64((365*24*time.Hour)/time.Second) {
		return errors.New("Telegram FloodWait 参数无效")
	}
	until := time.Now().Add(time.Duration(seconds) * time.Second)
	result := s.db.WithContext(ctx).Model(&model.TelegramAPIConfig{}).
		Where("api_id = ?", apiID).
		Updates(map[string]any{"flood_wait_until": until, "last_flood_wait_seconds": seconds})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func validateTelegramAPIConfig(input xyRequest.SaveTelegramAPIConfig, requireHash bool) error {
	if strings.TrimSpace(input.Name) == "" || len(strings.TrimSpace(input.Name)) > 100 || input.APIID <= 0 || input.APIID > 2147483647 {
		return errors.New("Telegram API 名称或 ID 无效")
	}
	apiHash := strings.TrimSpace(input.APIHash)
	if requireHash || apiHash != "" {
		decoded, err := hex.DecodeString(apiHash)
		if err != nil || len(decoded) != 16 {
			return errors.New("API Hash 必须是 32 位十六进制字符串")
		}
	}
	return nil
}

func telegramAPIHashAAD(id uint) string {
	return fmt.Sprintf("telegram-api:%d:api-hash", id)
}
