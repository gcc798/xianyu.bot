package service

import (
	"testing"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestDeliveryRulePrefersSKUAndChecksOwner 验证对应测试场景。
func TestDeliveryRulePrefersSKUAndChecksOwner(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Product{}, &model.DeliveryRule{}))

	account := model.Account{OwnerID: 10, Label: "测试账号", Status: "configured", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	product := model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-1", Title: "测试商品"}
	require.NoError(t, db.Create(&product).Error)

	service := &DeliveryRule{db: db}
	require.NoError(t, service.Save(t.Context(), 10, xyRequest.SaveDeliveryRule{
		ProductID: product.ID, Mode: model.DeliveryModeText, Template: "商品通用内容", Enabled: true,
	}))
	require.NoError(t, service.Save(t.Context(), 10, xyRequest.SaveDeliveryRule{
		ProductID: product.ID, SKUKey: "sku-red", SKUName: "红色", Mode: model.DeliveryModeCard,
		Template: "您的卡密：{card}", Enabled: true,
	}))
	require.NoError(t, service.Save(t.Context(), 10, xyRequest.SaveDeliveryRule{
		ProductID: product.ID, SKUKey: "disabled", Mode: model.DeliveryModeText, Template: "停用规则", Enabled: false,
	}))

	match, err := service.Resolve(t.Context(), 10, product.ID, "sku-red")
	require.NoError(t, err)
	require.Equal(t, "sku-red", match.SKUKey)
	require.Equal(t, model.DeliveryModeCard, match.Mode)
	match, err = service.Resolve(t.Context(), 10, product.ID, "sku-blue")
	require.NoError(t, err)
	require.Empty(t, match.SKUKey)
	require.Equal(t, "商品通用内容", match.Template)
	enabledRules, enabledTotal, err := service.List(t.Context(), 10, xyRequest.DeliveryRuleSearch{ProductID: product.ID, EnabledOnly: true})
	require.NoError(t, err)
	require.EqualValues(t, 2, enabledTotal)
	require.Len(t, enabledRules, 2)

	list, total, err := service.List(t.Context(), 11, xyRequest.DeliveryRuleSearch{})
	require.NoError(t, err)
	require.Zero(t, total)
	require.Empty(t, list)
	require.Error(t, service.Save(t.Context(), 11, xyRequest.SaveDeliveryRule{
		ProductID: product.ID, Mode: model.DeliveryModeText, Template: "越权", Enabled: true,
	}))
	require.Error(t, service.Delete(t.Context(), 11, 1))
	require.Error(t, service.Save(t.Context(), 10, xyRequest.SaveDeliveryRule{
		ProductID: product.ID, SKUKey: "invalid", Mode: model.DeliveryModeCard, Template: "缺少占位符", Enabled: true,
	}))
}
