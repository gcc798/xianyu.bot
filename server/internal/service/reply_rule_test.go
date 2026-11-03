package service

import (
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestReplyRuleResolveOrderAndOwnerIsolation 验证对应测试场景。
func TestReplyRuleResolveOrderAndOwnerIsolation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Product{}, &model.ReplyRule{}, &model.ReplyContent{}))

	account := model.Account{OwnerID: 10, Label: "测试账号", Status: "configured", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	product := model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-1", Title: "测试商品"}
	require.NoError(t, db.Create(&product).Error)

	service := &ReplyRule{db: db}
	for _, rule := range []xyRequest.SaveReplyRule{
		{ProductID: product.ID, Type: model.ReplyRuleFuzzy, Keyword: "怎么", Priority: 100, Enabled: true, Content: "模糊回复"},
		{ProductID: product.ID, Type: model.ReplyRuleFuzzy, Keyword: "卖", Priority: 200, Enabled: true, Content: "高优先级回复"},
		{ProductID: product.ID, Type: model.ReplyRuleFallback, Priority: 1000, Enabled: true, Content: "兜底回复"},
	} {
		require.NoError(t, service.Save(t.Context(), 10, rule))
	}
	legacyRule := model.ReplyRule{OwnerID: 10, ProductID: product.ID, Type: model.ReplyRuleExact, Keyword: "怎么卖", Priority: 1, Enabled: true}
	require.NoError(t, db.Create(&legacyRule).Error)
	require.NoError(t, db.Create(&model.ReplyContent{RuleID: legacyRule.ID, Text: "旧精确规则"}).Error)
	require.Error(t, service.Save(t.Context(), 10, xyRequest.SaveReplyRule{ProductID: product.ID, Type: model.ReplyRuleExact, Keyword: "不允许新增", Enabled: true, Content: "旧类型"}))
	require.Error(t, service.Save(t.Context(), 10, xyRequest.SaveReplyRule{ProductID: product.ID, Type: model.ReplyRuleFallback, Enabled: true, Content: "重复兜底"}))
	ownerList, ownerTotal, err := service.List(t.Context(), 10, xyRequest.ReplyRuleSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 4, ownerTotal)
	require.Len(t, ownerList, 4)
	require.Equal(t, model.ReplyRuleFallback, ownerList[len(ownerList)-1].Type)

	match, err := service.Resolve(t.Context(), 10, product.ID, "怎么卖")
	require.NoError(t, err)
	require.Equal(t, "高优先级回复", match.Content)
	match, err = service.Resolve(t.Context(), 10, product.ID, "这个怎么使用")
	require.NoError(t, err)
	require.Equal(t, "模糊回复", match.Content)
	match, err = service.Resolve(t.Context(), 10, product.ID, "其他问题")
	require.NoError(t, err)
	require.Equal(t, "兜底回复", match.Content)

	list, total, err := service.List(t.Context(), 11, xyRequest.ReplyRuleSearch{})
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, list)
	require.Error(t, service.Save(t.Context(), 11, xyRequest.SaveReplyRule{ProductID: product.ID, Type: model.ReplyRuleExact, Keyword: "越权", Enabled: true, Content: "越权"}))
	require.Error(t, service.Delete(t.Context(), 11, 1))
}
