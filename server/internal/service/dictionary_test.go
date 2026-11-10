package service

import (
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDictionaryOptionsReturnsEnabledCodeNames(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Dictionary{}, &model.DictionaryDetail{}))

	enabled, disabled := true, false
	status := model.Dictionary{Type: "order_status", Status: &enabled}
	require.NoError(t, db.Create(&status).Error)
	require.NoError(t, db.Create(&model.DictionaryDetail{
		DictionaryID: int(status.ID), Value: "pending_delivery", Label: "待发货", Status: &enabled, Sort: 1,
	}).Error)
	require.NoError(t, db.Create(&model.DictionaryDetail{
		DictionaryID: int(status.ID), Value: "legacy", Label: "旧值", Status: &disabled, Sort: 2,
	}).Error)
	ignored := model.Dictionary{Type: "inactive", Status: &disabled}
	require.NoError(t, db.Create(&ignored).Error)
	require.NoError(t, db.Create(&model.DictionaryDetail{
		DictionaryID: int(ignored.ID), Value: "x", Label: "不应返回", Status: &enabled,
	}).Error)

	options, err := NewDictionaryService(db).Options(t.Context())
	require.NoError(t, err)
	require.Equal(t, []response.DictionaryOption{{Code: "pending_delivery", Name: "待发货"}}, options["order_status"])
	require.NotContains(t, options, "inactive")
}
