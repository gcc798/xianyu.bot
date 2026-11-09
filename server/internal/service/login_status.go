package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"gorm.io/gorm"
)

// hasLoginEndpoint 是闲鱼 Web 登录态检查入口。
const hasLoginEndpoint = "https://passport.goofish.com/newlogin/hasLogin.do"

// credentialRefreshLocks 串行化同一账号的 Cookie 刷新与持久化。
var credentialRefreshLocks sync.Map

// LoginCheck 是一次远端登录态检查结果。
type LoginCheck struct {
	valid  bool   // valid 表示凭据仍可直接使用。
	risk   bool   // risk 表示需要用户完成安全验证。
	cookie string // cookie 是服务端更新后的完整 Cookie。
}

// loginChecker 描述账号服务依赖的登录态检查能力。
type loginChecker interface {
	// Check 使用完整 Cookie 检查远端登录态。
	Check(context.Context, string) (LoginCheck, error)
}

// HTTPLoginChecker 通过闲鱼 Passport HTTP 接口检查登录态。
type HTTPLoginChecker struct {
	client   *http.Client // client 执行带超时的登录态请求。
	endpoint string       // endpoint 是可在测试中替换的检查地址。
}

// NewHTTPLoginChecker 创建使用官方 Passport 入口的检查器。
func NewHTTPLoginChecker() *HTTPLoginChecker {
	return &HTTPLoginChecker{
		client:   &http.Client{Timeout: 20 * time.Second},
		endpoint: hasLoginEndpoint,
	}
}

// CheckLogin 检查 ownerID 名下账号并原子保存远端刷新后的 Cookie。
func (s *Account) CheckLogin(ctx context.Context, ownerID, accountID uint) (xyResponse.LoginStatus, error) {
	if ownerID == 0 || accountID == 0 {
		return xyResponse.LoginStatus{}, errors.New("账号 ID 不能为空")
	}
	// shortcut: account locks live for the process lifetime; replace if high account churn becomes a real workload.
	lockValue, _ := credentialRefreshLocks.LoadOrStore(accountID, new(sync.Mutex))
	lock := lockValue.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	db := s.db.WithContext(ctx)
	var account model.Account
	if err := db.Where("id = ? AND owner_id = ?", accountID, ownerID).First(&account).Error; err != nil {
		return xyResponse.LoginStatus{}, err
	}
	var credential model.Credential
	if err := db.Where("account_id = ?", accountID).First(&credential).Error; err != nil {
		return xyResponse.LoginStatus{}, err
	}
	cookie, err := decryptCredential(s.db, credential.CookieCiphertext, credentialAAD(accountID, "cookie"))
	if err != nil {
		return xyResponse.LoginStatus{}, err
	}
	originalCookie := cookie

	checker := s.loginChecker
	if checker == nil {
		checker = NewHTTPLoginChecker()
	}
	var check LoginCheck
	for attempt := range 2 {
		check, err = checker.Check(ctx, cookie)
		if check.cookie != "" {
			cookie = check.cookie
		}
		if err == nil && (check.valid || check.risk) {
			break
		}
		if attempt == 0 {
			if !waitContext(ctx, 500*time.Millisecond) {
				return xyResponse.LoginStatus{}, context.Cause(ctx)
			}
		}
	}
	if err != nil {
		return xyResponse.LoginStatus{}, err
	}

	status := "expired"
	if check.risk {
		status = "verification_required"
	} else if check.valid {
		status = "ready"
	}
	cookieUpdated := cookie != "" && cookie != originalCookie

	err = db.Transaction(func(tx *gorm.DB) error {
		if cookieUpdated {
			ciphertext, encryptErr := encryptCredential(s.db, cookie, credentialAAD(accountID, "cookie"))
			if encryptErr != nil {
				return encryptErr
			}
			result := tx.Model(&model.Credential{}).
				Where("id = ? AND account_id = ? AND version = ?", credential.ID, accountID, credential.Version).
				Updates(map[string]any{"cookie_ciphertext": ciphertext, "version": gorm.Expr("version + 1")})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return errors.New("凭证已被其他操作更新，请重试")
			}
			credential.Version++
		}
		return tx.Model(&model.Account{}).Where("id = ? AND owner_id = ?", accountID, ownerID).Update("status", status).Error
	})
	if err != nil {
		return xyResponse.LoginStatus{}, err
	}
	s.events.Record(ctx, model.BusinessEvent{OwnerID: ownerID, AccountID: accountID, Type: "login_check", Status: status, Summary: "闲鱼账号登录态检查完成"})
	return xyResponse.LoginStatus{Status: status, CookieUpdated: cookieUpdated, CredentialVersion: credential.Version}, nil
}

// Check 调用 Passport 接口并合并响应中的 Cookie 更新。
func (c *HTTPLoginChecker) Check(ctx context.Context, cookie string) (LoginCheck, error) {
	cookies := parseCookieHeader(cookie)
	form := url.Values{
		"appName":         {"xianyu"},
		"fromSite":        {"77"},
		"hid":             {cookies["unb"]},
		"ltl":             {"true"},
		"appEntrance":     {"web"},
		"_csrf_token":     {cookies["XSRF-TOKEN"]},
		"umidToken":       {""},
		"hsiz":            {cookies["cookie2"]},
		"bizParams":       {"taobaoBizLoginFrom=web"},
		"mainPage":        {"false"},
		"isMobile":        {"false"},
		"lang":            {"zh_CN"},
		"returnUrl":       {""},
		"isIframe":        {"true"},
		"documentReferer": {"https://www.goofish.com/"},
		"defaultView":     {"hasLogin"},
		"umidTag":         {"SERVER"},
		"deviceId":        {cookies["cna"]},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return LoginCheck{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json, text/plain, */*")
	request.Header.Set("Origin", "https://passport.goofish.com")
	request.Header.Set("Referer", "https://passport.goofish.com/")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	request.Header.Set("Cookie", cookie)

	response, err := c.client.Do(request)
	if err != nil {
		return LoginCheck{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return LoginCheck{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return LoginCheck{}, fmt.Errorf("hasLogin HTTP 状态 %d", response.StatusCode)
	}
	text := string(body)
	merged := cookie
	if responseCookies := response.Cookies(); len(responseCookies) > 0 {
		merged = mergeResponseCookies(cookies, responseCookies)
	}
	if strings.Contains(text, "RGV587_ERROR") || strings.Contains(text, "被挤爆啦") {
		return LoginCheck{risk: true, cookie: merged}, nil
	}
	var payload struct {
		Content struct {
			Success bool `json:"success"` // Success 表示远端确认已登录。
		} `json:"content"` // Content 包含 Passport 业务结果。
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return LoginCheck{}, errors.New("hasLogin 响应格式无法识别")
	}
	return LoginCheck{valid: payload.Content.Success, cookie: merged}, nil
}

// parseCookieHeader 将 Cookie 请求头解析为名称和值映射。
func parseCookieHeader(header string) map[string]string {
	cookies := make(map[string]string)
	for part := range strings.SplitSeq(header, ";") {
		name, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if ok && name != "" {
			cookies[name] = value
		}
	}
	return cookies
}

// mergeResponseCookies 应用响应 Cookie 的新增、更新和删除后稳定排序输出。
func mergeResponseCookies(current map[string]string, updates []*http.Cookie) string {
	for _, cookie := range updates {
		if cookie.MaxAge < 0 {
			delete(current, cookie.Name)
		} else if cookie.Name != "" {
			current[cookie.Name] = cookie.Value
		}
	}
	keys := slices.Sorted(maps.Keys(current))
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+current[key])
	}
	return strings.Join(parts, "; ")
}

// waitContext 等待 duration，ctx 取消时提前返回 false。
func waitContext(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
