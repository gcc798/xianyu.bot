package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/gcc798/xianyu.bot/internal/security"
	"github.com/gin-gonic/gin"
)

// TelegramProxyAPI 处理Telegram 代理 HTTP 请求。
type TelegramProxyAPI struct{}

// NewTelegramProxyAPI 创建 TelegramProxyAPI。
func NewTelegramProxyAPI() *TelegramProxyAPI { return new(TelegramProxyAPI) }

// telegramHTTPClient 执行允许长轮询的 Telegram 请求。
var telegramHTTPClient = &http.Client{Timeout: 15 * time.Minute}

// Proxy preserves the Telegram service method, path, query, status, headers, and body behind system authentication.
func (a *TelegramProxyAPI) Proxy(c *gin.Context) {
	upstream, err := telegramRequest(c)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	defer upstream.Body.Close()
	for name, values := range upstream.Header {
		if isHopByHopHeader(name) {
			continue
		}
		for _, value := range values {
			c.Writer.Header().Add(name, value)
		}
	}
	c.Status(upstream.StatusCode)
	_, _ = io.Copy(c.Writer, upstream.Body)
}

// telegramRequest 构造并发送 Telegram 上游请求。
func telegramRequest(c *gin.Context) (*http.Response, error) {
	base := strings.TrimRight(os.Getenv("TELEGRAM_SERVICE_URL"), "/")
	if base == "" {
		base = "http://telegram:8080"
	}
	parsed, err := url.Parse(base)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("Telegram 服务地址配置无效")
	}
	path := c.Param("path")
	if path == "" {
		_, path, _ = strings.Cut(c.Request.URL.Path, "/xy/proxy/telegram")
	}
	if path == "" || !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("Telegram 服务请求路径无效")
	}
	request, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, base+path+querySuffix(c), c.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("创建 Telegram 服务请求失败")
	}
	for _, name := range []string{"Accept", "Content-Type", "If-Modified-Since", "If-None-Match", "Range"} {
		if value := c.GetHeader(name); value != "" {
			request.Header.Set(name, value)
		}
	}
	request.Header.Set("X-Console-User-ID", fmt.Sprint(security.GetUserID(c)))
	upstream, err := telegramHTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("Telegram 服务不可用")
	}
	return upstream, nil
}

// querySuffix 返回保留查询参数的路径后缀。
func querySuffix(c *gin.Context) string {
	if c.Request.URL.RawQuery == "" {
		return ""
	}
	return "?" + c.Request.URL.RawQuery
}

// isHopByHopHeader 判断是否为逐跳 HTTP 头。
func isHopByHopHeader(name string) bool {
	switch http.CanonicalHeaderKey(name) {
	case "Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade":
		return true
	default:
		return false
	}
}
