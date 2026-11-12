package service

import (
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestBusinessEventListIsOwnerScoped 验证业务事件查询不会跨用户返回记录。
func TestBusinessEventListIsOwnerScoped(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Product{}, &model.Order{}, &model.BusinessEvent{}))
	require.NoError(t, db.Create(&[]model.BusinessEvent{
		{OwnerID: 10, Type: "delivery", Status: "failed", Summary: "订单发货失败"},
		{OwnerID: 11, Type: "delivery", Status: "failed", Summary: "其他用户事件"},
	}).Error)

	service := &BusinessEvent{db: db}
	events, total, err := service.List(t.Context(), 10, xyRequest.BusinessEventSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, events, 1)
	require.Equal(t, "订单发货失败", events[0].Summary)

	otherEvents, otherTotal, err := service.List(t.Context(), 12, xyRequest.BusinessEventSearch{})
	require.NoError(t, err)
	require.Zero(t, otherTotal)
	require.Empty(t, otherEvents)
}
