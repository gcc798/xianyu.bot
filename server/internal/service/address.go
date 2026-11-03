package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// amapInputTipsURL 是高德地点提示查询入口。
	amapInputTipsURL = "https://restapi.amap.com/v3/assistant/inputtips"
	// addressHTTPClient 为地点解析设置统一请求超时。
	addressHTTPClient = &http.Client{Timeout: 20 * time.Second}
)

// resolvePublishAddress 将用户地点关键词解析为闲鱼发布所需的行政区和坐标。
func resolvePublishAddress(keyword string) (map[string]any, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, errors.New("开启自动续售时请填写商品所在地")
	}
	query := url.Values{
		"s": {"rsv3"}, "key": {"c9b68d4ce9a2a97f22a4a439404488ca"}, "platform": {"JS"},
		"logversion": {"2.0"}, "sdkversion": {"2.0"}, "appname": {"https://seller.goofish.com"},
		"keywords": {keyword}, "city": {"全国"}, "citylimit": {"false"}, "datatype": {"all"},
	}
	req, err := http.NewRequest(http.MethodGet, amapInputTipsURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Origin", "https://seller.goofish.com")
	req.Header.Set("Referer", "https://seller.goofish.com/")
	req.Header.Set("User-Agent", webBrowserUA)
	resp, err := addressHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("解析商品所在地: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Status string           `json:"status"` // Status 是高德业务状态码。
		Info   string           `json:"info"`   // Info 是失败时的状态说明。
		Tips   []map[string]any `json:"tips"`   // Tips 是匹配的地点候选。
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析商品所在地响应: %w", err)
	}
	if result.Status != "1" {
		return nil, fmt.Errorf("解析商品所在地失败: %s", result.Info)
	}
	for _, tip := range result.Tips {
		id, name := amapText(tip["id"]), amapText(tip["name"])
		adcode, locationText := amapText(tip["adcode"]), amapText(tip["location"])
		location := strings.Split(locationText, ",")
		if id != "" && name != "" && adcode != "" && len(location) == 2 {
			return map[string]any{"divisionId": adcode, "gps": location[1] + "," + location[0], "poiId": id, "poiName": name}, nil
		}
	}
	return nil, errors.New("未找到有效的商品所在地，请填写更具体的地区")
}

// amapText 安全提取高德响应中的字符串字段。
func amapText(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}
