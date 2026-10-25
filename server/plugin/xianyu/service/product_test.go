package service

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model"
	xianyuRequest "github.com/flipped-aurora/gin-vue-admin/server/plugin/xianyu/model/request"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProductSyncUsesRemoteDataAndIsOwnerScoped(t *testing.T) {
	t.Setenv(credentialKeyEnvironment, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Credential{}, &model.Product{}, &model.ProductSetting{}))

	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = db
	account := model.Account{OwnerID: 10, Label: "测试账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	ciphertext, err := encryptCredential("_m_h5_tk=oldtoken_1; unb=12345", credentialAAD(account.ID, "cookie"))
	require.NoError(t, err)
	require.NoError(t, db.Create(&model.Credential{AccountID: account.ID, CookieCiphertext: ciphertext}).Error)
	require.NoError(t, db.Create(&model.Product{
		OwnerID: 10, AccountID: account.ID, ExternalProductID: "stale", Title: "旧商品", Status: "on_sale",
	}).Error)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "34839810", r.URL.Query().Get("appKey"))
		require.NoError(t, r.ParseForm())
		require.Contains(t, r.Form.Get("data"), `"pageNumber":1`)
		http.SetCookie(w, &http.Cookie{Name: "_m_h5_tk", Value: "newtoken_2"})
		_, _ = w.Write([]byte(`{
			"ret":["SUCCESS::调用成功"],
			"data":{"cardList":[{"cardData":{
				"title":"真实商品","priceInfo":{"price":"19.90"},"picInfo":{"picUrl":"https://img.example/item.jpg"},
				"detailParams":{"itemId":"item-1"}
			}}]}
		}`))
	}))
	t.Cleanup(server.Close)

	service := &product{client: server.Client(), endpoint: server.URL}
	input := xianyuRequest.ProductSync{AccountID: account.ID}
	_, err = service.Sync(t.Context(), 11, input)
	require.Error(t, err)
	count, err := service.Sync(t.Context(), 10, input)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	list, total, err := service.List(t.Context(), 10, xianyuRequest.ProductSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	require.Equal(t, "真实商品", list[0].Title)
	require.EqualValues(t, 1990, list[0].PriceCent)
	require.Equal(t, "offline", list[1].Status)

	var credential model.Credential
	require.NoError(t, db.Where("account_id = ?", account.ID).First(&credential).Error)
	require.EqualValues(t, 2, credential.Version)
	updatedCookie, err := decryptCredential(credential.CookieCiphertext, credentialAAD(account.ID, "cookie"))
	require.NoError(t, err)
	require.Contains(t, updatedCookie, "_m_h5_tk=newtoken_2")

	require.Error(t, service.UpdateSetting(t.Context(), 11, xianyuRequest.UpdateProductSetting{ProductID: list[0].ID, AutoReplyEnabled: true}))
	require.NoError(t, service.UpdateSetting(t.Context(), 10, xianyuRequest.UpdateProductSetting{ProductID: list[0].ID, AutoReplyEnabled: true, ReplyDelaySeconds: 5}))
}

func TestProductListFiltersByTitleOrIDAndStatus(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.Account{}, &model.Product{}, &model.ProductSetting{}))
	oldDB := global.GVA_DB
	t.Cleanup(func() { global.GVA_DB = oldDB })
	global.GVA_DB = db
	account := model.Account{OwnerID: 10, Label: "商品账号", Status: "ready", Enabled: true}
	require.NoError(t, db.Create(&account).Error)
	updatedAt := func(day int) time.Time { return time.Date(2026, time.January, day, 0, 0, 0, 0, time.UTC) }
	for _, item := range []model.Product{
		{GVA_MODEL: global.GVA_MODEL{UpdatedAt: updatedAt(1)}, OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-1", Title: "天涯书籍", Status: "on_sale", SourceJSON: `{"detailParams":{"imageInfos":"[{\"url\":\"https://img.example/1.jpg\"},{\"url\":\"https://img.example/2.jpg\"}]"}}`},
		{GVA_MODEL: global.GVA_MODEL{UpdatedAt: updatedAt(2)}, OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-2", Title: "天涯旧书", Status: "sold"},
		{GVA_MODEL: global.GVA_MODEL{UpdatedAt: updatedAt(3)}, OwnerID: 10, AccountID: account.ID, ExternalProductID: "item-3", Title: "其他商品", Status: "offline"},
	} {
		require.NoError(t, db.Create(&item).Error)
	}
	list, total, err := new(product).List(t.Context(), 10, xianyuRequest.ProductSearch{})
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	require.Equal(t, []string{"item-1", "item-3", "item-2"}, []string{list[0].ExternalProductID, list[1].ExternalProductID, list[2].ExternalProductID})
	require.Equal(t, []string{"https://img.example/1.jpg", "https://img.example/2.jpg"}, list[0].ImageURLs)
	list, total, err = new(product).List(t.Context(), 10, xianyuRequest.ProductSearch{TitleOrID: "天涯", Status: "sold"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "item-2", list[0].ExternalProductID)
	list, total, err = new(product).List(t.Context(), 10, xianyuRequest.ProductSearch{TitleOrID: "item-3"})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Equal(t, "item-3", list[0].ExternalProductID)
	_, _, err = new(product).List(t.Context(), 10, xianyuRequest.ProductSearch{Status: "deleted"})
	require.ErrorContains(t, err, "商品状态无效")
}

func TestProductImageURLsReadsSavedImageInfosAndFallsBackToCover(t *testing.T) {
	images := productImageURLs(`{"detailParams":{"imageInfos":"[{\"url\":\"https://img.example/1.jpg\"},{\"url\":\"https://img.example/2.jpg\"},{\"url\":\"https://img.example/1.jpg\"}]"}}`, "https://img.example/cover.jpg")
	require.Equal(t, []string{"https://img.example/1.jpg", "https://img.example/2.jpg"}, images)
	require.Equal(t, []string{"https://img.example/cover.jpg"}, productImageURLs("{}", "https://img.example/cover.jpg"))
	require.Nil(t, productImageURLs("{}", ""))
}

func TestParseProductPageRejectsFailedOrInvalidPrice(t *testing.T) {
	_, err := parseProductPage([]byte(`{"ret":["FAIL_SYS_TOKEN_EXOIRED::令牌过期"]}`))
	require.EqualError(t, err, "闲鱼商品同步失败：FAIL_SYS_TOKEN_EXOIRED")
	_, err = parseProductPage([]byte(`{
		"ret":["SUCCESS::ok"],"data":{"cardList":[{"cardData":{"id":"1","title":"商品","priceInfo":{"price":"free"}}}]}
	}`))
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "价格无效"))
}

func TestLocalPublishPayloadPreservesCategoryAndResolvesAddress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "西湖区", r.URL.Query().Get("keywords"))
		_, _ = w.Write([]byte(`{"status":"1","tips":[{"id":"poi-1","name":"西湖区","adcode":"330106","location":"120.13026,30.25961"},{"id":[],"name":"杭州市","adcode":"330100","location":[]}]}`))
	}))
	t.Cleanup(server.Close)
	oldURL, oldClient := amapInputTipsURL, addressHTTPClient
	t.Cleanup(func() { amapInputTipsURL, addressHTTPClient = oldURL, oldClient })
	amapInputTipsURL, addressHTTPClient = server.URL, server.Client()

	item := model.Product{
		Title: "测试商品", PriceCent: 100,
		SourceJSON: `{"categoryId":50023914,"detailParams":{"imageInfos":"[{\"url\":\"https://img.example/item.jpg\",\"major\":true,\"type\":0}]"}}`,
	}
	payload, err := localPublishPayload(item, xianyuRequest.UpdateProductSetting{RelistAddress: "西湖区"})
	require.NoError(t, err)
	require.Equal(t, "50023914", payload["itemCatDTO"].(map[string]any)["catId"])
	require.Equal(t, "30.25961,120.13026", payload["itemAddrDTO"].(map[string]any)["gps"])
}

func TestLocalPublishPayloadInheritsOriginalAddress(t *testing.T) {
	item := model.Product{
		Title: "测试商品", PriceCent: 100,
		SourceJSON:      `{"categoryId":50023914,"detailParams":{"imageInfos":"[{\"url\":\"https://img.example/item.jpg\",\"major\":true,\"type\":0}]"}}`,
		PublishSnapshot: `{"api":"mtop.idle.pc.idleitem.publish","payload":{"itemAddrDTO":{"divisionId":"330106","gps":"30.25961,120.13026","poiId":"poi-1","poiName":"西湖区"}}}`,
	}
	payload, err := localPublishPayload(item, xianyuRequest.UpdateProductSetting{})
	require.NoError(t, err)
	require.Equal(t, "西湖区", payload["itemAddrDTO"].(map[string]any)["poiName"])
}

func TestLocalPublishPayloadRejectsEmptyInheritedAddress(t *testing.T) {
	item := model.Product{
		Title: "测试商品", PriceCent: 100,
		SourceJSON:      `{"categoryId":50023914,"detailParams":{"imageInfos":"[{\"url\":\"https://img.example/item.jpg\",\"major\":true,\"type\":0}]"}}`,
		PublishSnapshot: `{"api":"mtop.idle.pc.idleitem.publish","payload":{"itemAddrDTO":{}}}`,
	}
	_, err := localPublishPayload(item, xianyuRequest.UpdateProductSetting{})
	require.EqualError(t, err, "无法从闲鱼读取原商品的发布所在地，请重新登录后重试")
}
