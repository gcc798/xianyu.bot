package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model/xianyu"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/xianyu/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/xianyu/response"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const productListEndpoint = "https://h5api.m.goofish.com/h5/mtop.idle.web.xyh.item.list/1.0/"

type product struct {
	db       *gorm.DB
	events   *businessEvent
	client   *http.Client
	endpoint string
	mtop     mtopCaller
}

type remoteProduct struct {
	ExternalProductID string
	Title             string
	PriceCent         int64
	ImageURL          string
	SourceJSON        string
}

type productCardData struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	PriceInfo struct {
		Price string `json:"price"`
	} `json:"priceInfo"`
	PicInfo struct {
		PicURL string `json:"picUrl"`
	} `json:"picInfo"`
	DetailParams struct {
		ItemID    string `json:"itemId"`
		Title     string `json:"title"`
		PicURL    string `json:"picUrl"`
		SoldPrice string `json:"soldPrice"`
	} `json:"detailParams"`
	Raw json.RawMessage `json:"-"`
}

func (c *productCardData) UnmarshalJSON(data []byte) error {
	type card productCardData
	if err := json.Unmarshal(data, (*card)(c)); err != nil {
		return err
	}
	c.Raw = append(c.Raw[:0], data...)
	return nil
}

func (s *product) List(ctx context.Context, ownerID uint, input xyRequest.ProductSearch) ([]xyResponse.Product, int64, error) {
	if ownerID == 0 {
		return nil, 0, errors.New("用户未登录")
	}
	db := s.db.WithContext(ctx).Table("xy_products AS p").Where("p.owner_id = ? AND p.deleted_at IS NULL", ownerID)
	if input.AccountID != 0 {
		db = db.Where("p.account_id = ?", input.AccountID)
	}
	if searchText := strings.TrimSpace(input.TitleOrID); searchText != "" {
		like := "%" + searchText + "%"
		db = db.Where("p.title LIKE ? OR p.external_product_id LIKE ?", like, like)
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		if status != "on_sale" && status != "offline" && status != "sold" {
			return nil, 0, errors.New("商品状态无效")
		}
		db = db.Where("p.status = ?", status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	list := make([]xyResponse.Product, 0)
	err := db.Select(`p.id, p.created_at, p.updated_at, p.account_id, a.label AS account_label,
		p.external_product_id, p.title, p.price_cent, p.status, p.image_url, p.source_json,
		COALESCE(s.virtual_product, false) AS virtual_product,
		COALESCE(s.auto_reply_enabled, false) AS auto_reply_enabled,
		COALESCE(s.auto_delivery_enabled, false) AS auto_delivery_enabled,
		COALESCE(s.auto_relist_enabled, false) AS auto_relist_enabled,
		COALESCE(s.reply_delay_seconds, 0) AS reply_delay_seconds,
		COALESCE(s.relist_delay_seconds, 60) AS relist_delay_seconds,
		COALESCE(s.relist_description, '') AS relist_description,
		COALESCE(s.relist_image_urls, '') AS relist_image_urls,
		COALESCE(s.relist_address, '') AS relist_address`).
		Joins("JOIN xy_accounts AS a ON a.id = p.account_id AND a.deleted_at IS NULL").
		Joins("LEFT JOIN xy_product_settings AS s ON s.product_id = p.id AND s.deleted_at IS NULL").
		Scopes(input.Paginate()).Order("CASE WHEN p.status = 'on_sale' THEN 0 ELSE 1 END, p.updated_at DESC, p.id DESC").Scan(&list).Error
	for index := range list {
		list[index].ImageURLs = productImageURLs(list[index].SourceJSON, list[index].ImageURL)
	}
	return list, total, err
}

func productImageURLs(sourceJSON, coverURL string) []string {
	var source struct {
		DetailParams struct {
			ImageInfos json.RawMessage `json:"imageInfos"`
		} `json:"detailParams"`
	}
	if err := json.Unmarshal([]byte(sourceJSON), &source); err == nil && len(source.DetailParams.ImageInfos) > 0 {
		imageInfos := source.DetailParams.ImageInfos
		if imageInfos[0] == '"' {
			var encoded string
			if json.Unmarshal(imageInfos, &encoded) == nil {
				imageInfos = []byte(encoded)
			}
		}
		var images []struct {
			URL string `json:"url"`
		}
		if json.Unmarshal(imageInfos, &images) == nil {
			urls := make([]string, 0, len(images))
			seen := make(map[string]struct{}, len(images))
			for _, image := range images {
				imageURL := strings.TrimSpace(image.URL)
				if imageURL == "" {
					continue
				}
				if _, exists := seen[imageURL]; exists {
					continue
				}
				seen[imageURL] = struct{}{}
				urls = append(urls, imageURL)
			}
			if len(urls) > 0 {
				return urls
			}
		}
	}
	if coverURL = strings.TrimSpace(coverURL); coverURL != "" {
		return []string{coverURL}
	}
	return nil
}

func (s *product) Sync(ctx context.Context, ownerID uint, input xyRequest.ProductSync) (count int, err error) {
	if ownerID == 0 || input.AccountID == 0 {
		return 0, errors.New("账号 ID 不能为空")
	}
	lockValue, _ := credentialRefreshLocks.LoadOrStore(input.AccountID, new(sync.Mutex))
	lock := lockValue.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	db := s.db.WithContext(ctx)
	var account model.Account
	if err := db.Where("id = ? AND owner_id = ?", input.AccountID, ownerID).First(&account).Error; err != nil {
		return 0, err
	}
	var credential model.Credential
	if err := db.Where("account_id = ?", input.AccountID).First(&credential).Error; err != nil {
		return 0, err
	}
	cookie, err := decryptCredential(s.db, credential.CookieCiphertext, credentialAAD(input.AccountID, "cookie"))
	if err != nil {
		return 0, err
	}
	originalCookie := cookie
	defer func() {
		if cookie == originalCookie {
			return
		}
		ciphertext, encryptErr := encryptCredential(s.db, cookie, credentialAAD(input.AccountID, "cookie"))
		if encryptErr == nil {
			result := db.Model(&model.Credential{}).
				Where("id = ? AND account_id = ? AND version = ?", credential.ID, input.AccountID, credential.Version).
				Updates(map[string]any{"cookie_ciphertext": ciphertext, "version": gorm.Expr("version + 1")})
			encryptErr = result.Error
			if encryptErr == nil && result.RowsAffected != 1 {
				encryptErr = errors.New("凭证已被其他操作更新，请重试")
			}
		}
		if err == nil && encryptErr != nil {
			err = encryptErr
		}
	}()

	all := make([]remoteProduct, 0)
	for page := 1; page <= 100; page++ {
		var items []remoteProduct
		var nextCookie string
		items, nextCookie, err = s.fetchPage(ctx, cookie, page, 20)
		if err != nil && nextCookie != cookie {
			cookie = nextCookie
			items, nextCookie, err = s.fetchPage(ctx, cookie, page, 20)
		}
		cookie = nextCookie
		if err != nil {
			return 0, err
		}
		all = append(all, items...)
		if len(items) < 20 {
			break
		}
		if page == 100 {
			return 0, errors.New("商品超过 2000 个，已停止同步")
		}
		if !waitContext(ctx, 300*time.Millisecond) {
			return 0, context.Cause(ctx)
		}
	}

	seen := make([]string, 0, len(all))
	products := make([]model.Product, 0, len(all))
	for _, source := range all {
		seen = append(seen, source.ExternalProductID)
		products = append(products, model.Product{
			OwnerID: ownerID, AccountID: input.AccountID, ExternalProductID: source.ExternalProductID,
			Title: source.Title, PriceCent: source.PriceCent, Status: "on_sale", ImageURL: source.ImageURL, SourceJSON: source.SourceJSON,
		})
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if len(products) > 0 {
			if createErr := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "account_id"}, {Name: "external_product_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"title", "price_cent", "status", "image_url", "source_json", "updated_at"}),
			}).Create(&products).Error; createErr != nil {
				return createErr
			}
		}
		offline := tx.Model(&model.Product{}).
			Where("owner_id = ? AND account_id = ? AND status = ?", ownerID, input.AccountID, "on_sale")
		if len(seen) > 0 {
			offline = offline.Where("external_product_id NOT IN ?", seen)
		}
		return offline.Update("status", "offline").Error
	})
	if err == nil {
		s.events.Record(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: input.AccountID, Type: "product_sync", Status: "success", Summary: fmt.Sprintf("同步 %d 个在售商品", len(all))})
	}
	return len(all), err
}

func (s *product) fetchPage(ctx context.Context, cookie string, page, pageSize int) ([]remoteProduct, string, error) {
	cookies := parseCookieHeader(cookie)
	token := strings.Split(cookies["_m_h5_tk"], "_")[0]
	if token == "" || cookies["unb"] == "" {
		return nil, cookie, errors.New("闲鱼凭证不完整，请重新扫码登录")
	}
	data, err := json.Marshal(map[string]any{
		"needGroupInfo": false, "pageNumber": page, "pageSize": pageSize,
		"groupName": "在售", "groupId": "58877261", "defaultGroup": true, "userId": cookies["unb"],
	})
	if err != nil {
		return nil, cookie, err
	}
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	hash := md5.Sum([]byte(token + "&" + timestamp + "&" + mtopAppKey + "&" + string(data)))
	target, err := url.Parse(s.endpointURL())
	if err != nil {
		return nil, cookie, err
	}
	query := target.Query()
	for key, value := range map[string]string{
		"jsv": "2.7.2", "appKey": mtopAppKey, "t": timestamp, "sign": hex.EncodeToString(hash[:]),
		"v": "1.0", "type": "originaljson", "accountSite": "xianyu", "dataType": "json",
		"timeout": "20000", "api": "mtop.idle.web.xyh.item.list", "sessionOption": "AutoLoginOnly",
		"spm_cnt": "a21ybx.im.0.0", "spm_pre": "a21ybx.collection.menu.1.272b5141NafCNK",
	} {
		query.Set(key, value)
	}
	target.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), strings.NewReader(url.Values{"data": {string(data)}}.Encode()))
	if err != nil {
		return nil, cookie, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Origin", "https://www.goofish.com")
	request.Header.Set("Referer", "https://www.goofish.com/")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/133.0.0.0 Safari/537.36")
	request.Header.Set("Cookie", cookie)
	response, err := s.httpClient().Do(request)
	if err != nil {
		return nil, cookie, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	mergedCookie := mergeResponseCookies(cookies, response.Cookies())
	if err != nil {
		return nil, mergedCookie, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, mergedCookie, fmt.Errorf("闲鱼商品接口 HTTP 状态 %d", response.StatusCode)
	}
	items, err := parseProductPage(body)
	return items, mergedCookie, err
}

func (s *product) httpClient() *http.Client {
	if s.client != nil {
		return s.client
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (s *product) endpointURL() string {
	if s.endpoint != "" {
		return s.endpoint
	}
	return productListEndpoint
}

func parseProductPage(body []byte) ([]remoteProduct, error) {
	var payload struct {
		Ret  []string `json:"ret"`
		Data struct {
			CardList []struct {
				CardData productCardData `json:"cardData"`
			} `json:"cardList"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, errors.New("闲鱼商品响应格式无法识别")
	}
	if len(payload.Ret) == 0 || !strings.Contains(payload.Ret[0], "SUCCESS") {
		if len(payload.Ret) > 0 {
			code, _, _ := strings.Cut(payload.Ret[0], "::")
			return nil, fmt.Errorf("闲鱼商品同步失败：%s", code)
		}
		return nil, errors.New("闲鱼商品同步失败：响应缺少状态")
	}
	items := make([]remoteProduct, 0, len(payload.Data.CardList))
	seen := make(map[string]struct{}, len(payload.Data.CardList))
	for _, card := range payload.Data.CardList {
		itemID := strings.TrimSpace(card.CardData.DetailParams.ItemID)
		if itemID == "" {
			itemID = strings.TrimSpace(card.CardData.ID)
		}
		title := strings.TrimSpace(card.CardData.Title)
		if title == "" {
			title = strings.TrimSpace(card.CardData.DetailParams.Title)
		}
		if itemID == "" || title == "" {
			continue
		}
		if _, exists := seen[itemID]; exists {
			continue
		}
		seen[itemID] = struct{}{}
		priceText := card.CardData.PriceInfo.Price
		if priceText == "" {
			priceText = card.CardData.DetailParams.SoldPrice
		}
		price, parseErr := strconv.ParseFloat(strings.TrimSpace(priceText), 64)
		if parseErr != nil || price < 0 {
			return nil, fmt.Errorf("商品 %s 的价格无效", itemID)
		}
		imageURL := strings.TrimSpace(card.CardData.PicInfo.PicURL)
		if imageURL == "" {
			imageURL = strings.TrimSpace(card.CardData.DetailParams.PicURL)
		}
		items = append(items, remoteProduct{
			ExternalProductID: itemID, Title: title, PriceCent: int64(math.Round(price * 100)), ImageURL: imageURL, SourceJSON: string(card.CardData.Raw),
		})
	}
	return items, nil
}

func (s *product) UpdateSetting(ctx context.Context, ownerID uint, input xyRequest.UpdateProductSetting) error {
	if ownerID == 0 || input.ProductID == 0 {
		return errors.New("商品 ID 不能为空")
	}
	if input.ReplyDelaySeconds < 0 || input.ReplyDelaySeconds > 3600 {
		return errors.New("回复延迟必须在 0 到 3600 秒之间")
	}
	if input.RelistDelaySeconds < 0 || input.RelistDelaySeconds > 86400 {
		return errors.New("自动续售延迟必须在 0 到 86400 秒之间")
	}
	db := s.db.WithContext(ctx)
	var item model.Product
	if err := db.Where("id = ? AND owner_id = ?", input.ProductID, ownerID).First(&item).Error; err != nil {
		return err
	}
	if strings.TrimSpace(input.RelistDescription) == item.Title {
		input.RelistDescription = ""
	}
	if input.AutoRelistEnabled {
		snapshot, err := s.capturePublishSnapshot(ctx, ownerID, item, input)
		if err != nil {
			return err
		}
		if strings.TrimSpace(input.RelistDescription) == "" {
			input.RelistDescription = publishSnapshotDescription(snapshot)
		}
		item.PublishSnapshot = snapshot
		if err := db.Model(&item).Update("publish_snapshot", snapshot).Error; err != nil {
			return err
		}
	}
	setting := model.ProductSetting{
		ProductID: input.ProductID, VirtualProduct: input.VirtualProduct, AutoReplyEnabled: input.AutoReplyEnabled,
		AutoDeliveryEnabled: input.AutoDeliveryEnabled, AutoRelistEnabled: input.AutoRelistEnabled,
		ReplyDelaySeconds: input.ReplyDelaySeconds, RelistDelaySeconds: input.RelistDelaySeconds,
		RelistDescription: strings.TrimSpace(input.RelistDescription), RelistImageURLs: strings.TrimSpace(input.RelistImageURLs),
		RelistAddress: strings.TrimSpace(input.RelistAddress),
	}
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "product_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"virtual_product": input.VirtualProduct, "auto_reply_enabled": input.AutoReplyEnabled,
			"auto_delivery_enabled": input.AutoDeliveryEnabled,
			"auto_relist_enabled":   input.AutoRelistEnabled, "reply_delay_seconds": input.ReplyDelaySeconds,
			"relist_delay_seconds": input.RelistDelaySeconds, "relist_description": strings.TrimSpace(input.RelistDescription),
			"relist_image_urls": strings.TrimSpace(input.RelistImageURLs), "relist_address": strings.TrimSpace(input.RelistAddress),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(&setting).Error
}
