package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRelistPublishesOnceAndMovesProductRules(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Product{}, &model.ProductSetting{}, &model.DeliveryRule{}, &model.ReplyRule{}, &model.Order{}, &model.RelistTask{}, &model.BusinessEvent{}))
	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = db

	account := model.Account{OwnerID: 10, Label: "续售账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	snapshot, err := json.Marshal(publishSnapshot{API: "mtop.idle.pc.idleitem.publish", Payload: map[string]any{
		"itemTextDTO": map[string]any{"title": "商品", "desc": "商品描述"},
		"itemCatDTO":  map[string]any{"catId": "cat-1", "channelCatId": "", "tbCatId": ""},
	}})
	require.NoError(t, err)
	item := model.Product{OwnerID: 10, AccountID: account.ID, ExternalProductID: "old-item", Title: "商品", Status: "sold", PublishSnapshot: string(snapshot)}
	require.NoError(t, db.Create(&item).Error)
	require.NoError(t, db.Create(&model.ProductSetting{ProductID: item.ID, AutoRelistEnabled: true}).Error)
	require.NoError(t, db.Create(&model.DeliveryRule{OwnerID: 10, ProductID: item.ID, Mode: model.DeliveryModeText, Template: "link", Enabled: true}).Error)
	order := model.Order{OwnerID: 10, AccountID: account.ID, ProductID: item.ID, ExternalOrderID: "order-1", Status: model.OrderStatusShipped}
	require.NoError(t, db.Create(&order).Error)
	publisher := &recordingMTopCaller{}
	publisher.calls = nil
	service := &relist{
		mtop:  mtopResultCaller{recordingMTopCaller: publisher, data: json.RawMessage(`{"itemId":"new-item"}`)},
		media: func(context.Context, uint, uint, map[string]any) error { return nil },
	}

	require.NoError(t, service.Plan(t.Context(), order))
	var task model.RelistTask
	require.NoError(t, db.First(&task).Error)
	require.NoError(t, db.Model(&task).Update("next_attempt_at", time.Now().Add(-time.Second)).Error)
	require.NoError(t, service.ProcessTask(t.Context(), 10, task.ID))
	require.NoError(t, db.First(&task, task.ID).Error)
	require.Equal(t, model.RelistTaskSucceeded, task.Status)
	require.Equal(t, "new-item", task.NewExternalProductID)
	var newItem model.Product
	require.NoError(t, db.Where("external_product_id = ?", "new-item").First(&newItem).Error)
	var rule model.DeliveryRule
	require.NoError(t, db.First(&rule).Error)
	require.Equal(t, newItem.ID, rule.ProductID)
	require.Equal(t, "mtop.taobao.idle.kgraph.pc.property.recommend/2.0", publisher.calls[0].api)
	category := publisher.calls[1].data.(map[string]any)["itemCatDTO"].(map[string]any)
	require.Equal(t, "channel-1", category["channelCatId"])
	require.Equal(t, "tb-1", category["tbCatId"])
}

type mtopResultCaller struct {
	*recordingMTopCaller
	data json.RawMessage
}

func (f mtopResultCaller) Call(ctx context.Context, ownerID, accountID uint, api string, data any, seller bool) (json.RawMessage, error) {
	f.recordingMTopCaller.Call(ctx, ownerID, accountID, api, data, seller)
	if strings.HasSuffix(api, "/2.0") {
		return json.RawMessage(`{"data":{"categoryPredictResult":{"catId":"cat-1","catName":"分类","channelCatId":"channel-1","tbCatId":"tb-1"}}}`), nil
	}
	return f.data, nil
}

func TestFindRecommendedCategoryRequiresExactOriginalCategory(t *testing.T) {
	fields, err := findRecommendedCategory(json.RawMessage(`{"data":{"cardList":[{"valuesList":[{"catId":"cat-1","channelCatId":"channel-1","tbCatId":"tb-1"}]}]}}`), "cat-1")
	require.NoError(t, err)
	require.Equal(t, "channel-1", fields["channelCatId"])
	_, err = findRecommendedCategory(json.RawMessage(`{"data":{"categoryPredictResult":{"catId":"other","channelCatId":"channel-1","tbCatId":"tb-1"}}}`), "cat-1")
	require.Error(t, err)
}
