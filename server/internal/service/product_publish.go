package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
)

// PublishSnapshot 保存自动重发商品所需的 MTop API 与完整负载。
type PublishSnapshot struct {
	API     string         `json:"api"`     // API 是发布商品使用的 MTop 方法。
	Seller  bool           `json:"seller"`  // Seller 决定使用卖家端请求参数。
	Payload map[string]any `json:"payload"` // Payload 是经过清理的商品发布参数。
}

// capturePublishSnapshot 从闲鱼详情或本地快照生成可重放的商品发布参数。
func (s *Product) capturePublishSnapshot(ctx context.Context, ownerID uint, item model.Product, input xyRequest.UpdateProductSetting) (string, error) {
	caller := s.mtop
	if caller == nil {
		caller = &MTopClient{db: s.db}
	}
	data, err := caller.Call(ctx, ownerID, item.AccountID, "mtop.idle.pc.backend.idleitem.editdetail", map[string]any{"itemId": item.ExternalProductID}, true)
	if err != nil {
		for _, api := range []string{"mtop.taobao.idle.pc.detail", "mtop.taobao.idle.item.pc.detail", "mtop.taobao.idle.item.detail"} {
			data, err = caller.Call(ctx, ownerID, item.AccountID, api, map[string]any{"itemId": item.ExternalProductID}, false)
			if err == nil {
				break
			}
		}
		if err != nil {
			payload, err := localPublishPayload(item, input)
			if err != nil {
				return "", fmt.Errorf("获取商品发布快照失败，本地商品信息也不完整: %w", err)
			}
			return marshalPublishSnapshot(payload)
		}
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", errors.New("闲鱼商品发布快照格式无法识别")
	}
	if nested, ok := payload["data"].(map[string]any); ok {
		payload = nested
	}
	detailPayload := payload
	for _, field := range []string{"itemTextDTO", "itemPriceDTO", "imageInfoDOList", "itemCatDTO", "itemAddrDTO", "itemPostFeeDTO"} {
		if payload[field] == nil {
			fallback, fallbackErr := localPublishPayload(item, input)
			if fallbackErr != nil {
				return "", fmt.Errorf("闲鱼商品详情不能直接用于续售（顶层字段: %s），本地商品信息也不完整: %w", strings.Join(sortedMapKeys(payload), ","), fallbackErr)
			}
			if strings.TrimSpace(input.RelistImageURLs) == "" {
				if images := detailPublishImages(detailPayload); len(images) > 0 {
					fallback["imageInfoDOList"] = images
				}
			}
			return marshalPublishSnapshot(fallback)
		}
	}
	if address, ok := payload["itemAddrDTO"].(map[string]any); !ok || !validPublishAddress(address) {
		fallback, fallbackErr := localPublishPayload(item, input)
		if fallbackErr != nil {
			return "", fmt.Errorf("闲鱼未返回原商品的发布所在地: %w", fallbackErr)
		}
		return marshalPublishSnapshot(fallback)
	}
	delete(payload, "itemId")
	delete(payload, "itemSkuList")
	delete(payload, "propertyImageList")
	payload["freebies"] = false
	payload["itemTypeStr"] = "b"
	payload["quantity"] = "1"
	payload["simpleItem"] = "true"
	payload["defaultPrice"] = false
	payload["sourceId"] = "pcMainPublish"
	payload["bizcode"] = "pcMainPublish"
	payload["publishScene"] = "pcMainPublish"
	if description := strings.TrimSpace(input.RelistDescription); description != "" {
		text, ok := payload["itemTextDTO"].(map[string]any)
		if !ok {
			return "", errors.New("闲鱼商品发布快照的文本格式无法识别")
		}
		text["desc"] = description
	}
	if images := configuredPublishImages(input.RelistImageURLs); len(images) > 0 {
		payload["imageInfoDOList"] = images
	}
	return marshalPublishSnapshot(payload)
}

// marshalPublishSnapshot 将发布负载包装为固定 API 的持久化快照。
func marshalPublishSnapshot(payload map[string]any) (string, error) {
	snapshot, err := json.Marshal(PublishSnapshot{API: "mtop.idle.pc.idleitem.publish", Payload: payload})
	if err != nil {
		return "", err
	}
	return string(snapshot), nil
}

// publishSnapshotDescription 提取已有快照中的商品描述。
func publishSnapshotDescription(raw string) string {
	var snapshot PublishSnapshot
	if json.Unmarshal([]byte(raw), &snapshot) != nil {
		return ""
	}
	text, _ := snapshot.Payload["itemTextDTO"].(map[string]any)
	description, _ := text["desc"].(string)
	return strings.TrimSpace(description)
}

// localPublishPayload 使用本地商品快照和用户设置补齐最小发布负载。
func localPublishPayload(item model.Product, input xyRequest.UpdateProductSetting) (map[string]any, error) {
	description := strings.TrimSpace(input.RelistDescription)
	if description == "" {
		// shortcut: 在售列表不返回正文，当详情接口被风控时允许用户编辑，未填则回退标题。
		description = item.Title
	}
	var source struct {
		CategoryID   json.RawMessage `json:"categoryId"` // CategoryID 是闲鱼平台类目。
		DetailParams struct {
			ImageInfos string `json:"imageInfos"` // ImageInfos 是原商品图片列表的序列化值。
			PostInfo   string `json:"postInfo"`   // PostInfo 是原商品发布附加信息。
		} `json:"detailParams"` // DetailParams 保存列表接口携带的详情字段。
	}
	if json.Unmarshal([]byte(item.SourceJSON), &source) != nil {
		return nil, errors.New("请先同步闲鱼商品")
	}
	categoryID := strings.Trim(strings.TrimSpace(string(source.CategoryID)), `"`)
	if categoryID == "" || categoryID == "null" {
		return nil, errors.New("续售商品缺少平台类目")
	}
	address := inheritedPublishAddress(item.PublishSnapshot)
	if address == nil && strings.TrimSpace(input.RelistAddress) != "" {
		var err error
		address, err = resolvePublishAddress(input.RelistAddress)
		if err != nil {
			return nil, err
		}
	}
	if address == nil {
		return nil, errors.New("无法从闲鱼读取原商品的发布所在地，请重新登录后重试")
	}
	images := configuredPublishImages(input.RelistImageURLs)
	if len(images) == 0 && source.DetailParams.ImageInfos != "" {
		_ = json.Unmarshal([]byte(source.DetailParams.ImageInfos), &images)
	}
	if len(images) == 0 && item.ImageURL != "" {
		images = append(images, map[string]any{"url": item.ImageURL, "major": true, "type": 0, "status": "done"})
	}
	if len(images) == 0 {
		return nil, errors.New("续售商品至少需要一张图片")
	}
	return map[string]any{
		"freebies": false, "itemTypeStr": "b", "quantity": "1", "simpleItem": "true",
		"imageInfoDOList":     images,
		"itemTextDTO":         map[string]any{"title": item.Title, "desc": description, "titleDescSeparate": true},
		"itemLabelExtList":    []any{},
		"itemPriceDTO":        map[string]any{"origPriceInCent": strconv.FormatInt(item.PriceCent, 10), "priceInCent": strconv.FormatInt(item.PriceCent, 10)},
		"userRightsProtocols": []any{},
		"itemPostFeeDTO":      map[string]any{"canFreeShipping": true, "supportFreight": true, "onlyTakeSelf": false},
		"itemAddrDTO":         address, "defaultPrice": false,
		"itemCatDTO": map[string]any{"catId": categoryID, "catName": "", "channelCatId": "", "tbCatId": ""},
		"sourceId":   "pcMainPublish", "bizcode": "pcMainPublish", "publishScene": "pcMainPublish",
	}, nil
}

// configuredPublishImages 将换行分隔地址转换为最多九张发布图片。
func configuredPublishImages(rawURLs string) []any {
	images := make([]any, 0, 9)
	for line := range strings.SplitSeq(strings.TrimSpace(rawURLs), "\n") {
		imageURL := strings.TrimSpace(line)
		if imageURL == "" {
			continue
		}
		images = append(images, map[string]any{"url": imageURL, "major": len(images) == 0, "type": 0, "status": "done"})
		if len(images) == 9 {
			break
		}
	}
	return images
}

// detailPublishImages 从商品详情响应提取最多九张发布图片。
func detailPublishImages(payload map[string]any) []any {
	item, ok := payload["itemDO"].(map[string]any)
	if !ok {
		return nil
	}
	values, ok := item["imageInfos"].([]any)
	if !ok {
		return nil
	}
	images := make([]any, 0, min(len(values), 9))
	for _, value := range values {
		image, ok := value.(map[string]any)
		if !ok {
			continue
		}
		imageURL := strings.TrimSpace(fmt.Sprint(image["url"]))
		if imageURL == "" || imageURL == "<nil>" {
			continue
		}
		images = append(images, map[string]any{"url": imageURL, "major": len(images) == 0, "type": 0, "status": "done"})
		if len(images) == 9 {
			break
		}
	}
	return images
}

// inheritedPublishAddress 从已有快照提取完整有效的发布所在地。
func inheritedPublishAddress(snapshot string) map[string]any {
	var saved PublishSnapshot
	if json.Unmarshal([]byte(snapshot), &saved) != nil {
		return nil
	}
	address, ok := saved.Payload["itemAddrDTO"].(map[string]any)
	if !ok {
		return nil
	}
	if !validPublishAddress(address) {
		return nil
	}
	return address
}

// validPublishAddress 校验闲鱼发布要求的行政区、坐标和地点字段。
func validPublishAddress(address map[string]any) bool {
	for _, key := range []string{"divisionId", "gps", "poiId", "poiName"} {
		value, exists := address[key]
		if !exists || value == nil || strings.TrimSpace(fmt.Sprint(value)) == "" {
			return false
		}
	}
	return true
}

// sortedMapKeys 返回稳定排序的映射键，用于可复现错误信息。
func sortedMapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
