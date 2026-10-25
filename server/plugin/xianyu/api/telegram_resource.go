package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type telegramResource struct{}

var telegramHTTPClient = &http.Client{Timeout: 15 * time.Minute}

func (a *telegramResource) ProxyJSON(c *gin.Context) {
	upstream, err := telegramRequest(c)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	defer upstream.Body.Close()
	var data any
	if err = json.NewDecoder(upstream.Body).Decode(&data); err != nil {
		response.FailWithMessage("Telegram 服务返回无效响应", c)
		return
	}
	if upstream.StatusCode >= http.StatusBadRequest {
		if detail, ok := data.(map[string]any)["detail"].(string); ok {
			response.FailWithMessage(detail, c)
			return
		}
		response.FailWithMessage("Telegram 服务请求失败", c)
		return
	}
	response.OkWithData(data, c)
}

func (a *telegramResource) ProxyFile(c *gin.Context) {
	upstream, err := telegramRequest(c)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	defer upstream.Body.Close()
	if upstream.StatusCode >= http.StatusBadRequest {
		response.FailWithMessage("文件不存在或暂不可用", c)
		return
	}
	for _, name := range []string{"Content-Type", "Content-Disposition", "Content-Length", "Accept-Ranges"} {
		if value := upstream.Header.Get(name); value != "" {
			c.Header(name, value)
		}
	}
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, upstream.Body)
}

func telegramRequest(c *gin.Context) (*http.Response, error) {
	base := strings.TrimRight(os.Getenv("TELEGRAM_SERVICE_URL"), "/")
	if base == "" {
		base = "http://telegram:8080"
	}
	parsed, err := url.Parse(base)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("Telegram 服务地址配置无效")
	}
	marker := "/xianyu/telegram"
	index := strings.Index(c.Request.URL.Path, marker)
	if index < 0 {
		return nil, fmt.Errorf("Telegram 服务请求路径无效")
	}
	path := c.Request.URL.Path[index+len(marker):]
	request, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, base+"/api/resources"+path+querySuffix(c), c.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("创建 Telegram 服务请求失败")
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Internal-Token", os.Getenv("TELEGRAM_INTERNAL_TOKEN"))
	upstream, err := telegramHTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Telegram 服务不可用")
	}
	return upstream, nil
}

func querySuffix(c *gin.Context) string {
	if c.Request.URL.RawQuery == "" {
		return ""
	}
	return "?" + c.Request.URL.RawQuery
}
