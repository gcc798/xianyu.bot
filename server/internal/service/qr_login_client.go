package service

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

// stringOrNumber 兼容闲鱼接口在字符串和数字间漂移的标识字段。
type stringOrNumber string

// UnmarshalJSON 将 JSON 字符串或数字统一解码为文本。
func (value *stringOrNumber) UnmarshalJSON(data []byte) error {
	var text string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
	} else {
		var number json.Number
		if err := json.Unmarshal(data, &number); err != nil {
			return errors.New("expected string or number")
		}
		text = number.String()
	}
	*value = stringOrNumber(text)
	return nil
}

// prepare 初始化 Cookie、读取登录表单并生成二维码图片。
func (m *QRLoginManager) prepare(ctx context.Context, session *QRSession) error {
	if err := m.refreshH5Token(ctx, session); err != nil {
		return fmt.Errorf("初始化二维码登录凭证失败: %w", err)
	}
	params, err := m.loginParams(ctx, session)
	if err != nil {
		return fmt.Errorf("获取二维码登录参数失败: %w", err)
	}
	for key, values := range params {
		session.params[key] = values
	}
	requestURL, err := url.Parse(m.endpoints.generate)
	if err != nil {
		return err
	}
	requestURL.RawQuery = session.params.Encode()
	request, err := NewQRRequest(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return err
	}
	var payload struct {
		Content struct {
			Success bool `json:"success"` // Success 表示二维码生成成功。
			Data    struct {
				T           stringOrNumber `json:"t"`           // T 是轮询时间参数。
				CK          stringOrNumber `json:"ck"`          // CK 是二维码校验参数。
				CodeContent string         `json:"codeContent"` // CodeContent 是二维码原始内容。
			} `json:"data"` // Data 包含二维码与轮询参数。
		} `json:"content"` // Content 包含二维码业务结果。
	}
	if err := doJSON(session.client, request, &payload); err != nil {
		return fmt.Errorf("生成二维码: %w", err)
	}
	if !payload.Content.Success || payload.Content.Data.CodeContent == "" {
		return errors.New("闲鱼未返回可用二维码")
	}
	session.params.Set("t", string(payload.Content.Data.T))
	session.params.Set("ck", string(payload.Content.Data.CK))
	png, err := qrcode.Encode(payload.Content.Data.CodeContent, qrcode.Medium, 320)
	if err != nil {
		return err
	}
	session.qrCodeDataURL = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
	return nil
}

// refreshH5Token 访问并签名 H5 首页接口，使会话 Cookie 包含有效 MTop 令牌。
func (m *QRLoginManager) refreshH5Token(ctx context.Context, session *QRSession) error {
	request, err := NewQRRequest(ctx, http.MethodGet, m.endpoints.h5Token, nil)
	if err != nil {
		return err
	}
	response, err := session.client.Do(request)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
	response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return fmt.Errorf("初始化 Cookie HTTP 状态 %d", response.StatusCode)
	}
	token := strings.Split(cookieValue(session.client, m.endpoints.h5Token, "_m_h5_tk"), "_")[0]
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	data := `{"bizScene":"home"}`
	hash := md5.Sum([]byte(token + "&" + timestamp + "&34839810&" + data))
	signedURL, err := url.Parse(m.endpoints.h5Token)
	if err != nil {
		return err
	}
	query := signedURL.Query()
	for key, value := range map[string]string{
		"jsv": "2.7.2", "appKey": "34839810", "t": timestamp,
		"sign": hex.EncodeToString(hash[:]), "v": "1.0", "type": "originaljson",
		"dataType": "json", "timeout": "20000", "api": "mtop.gaia.nodejs.gaia.idle.data.gw.v2.index.get", "data": data,
	} {
		query.Set(key, value)
	}
	signedURL.RawQuery = query.Encode()
	request, err = NewQRRequest(ctx, http.MethodPost, signedURL.String(), strings.NewReader(""))
	if err != nil {
		return err
	}
	response, err = session.client.Do(request)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
	response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return fmt.Errorf("刷新 Cookie HTTP 状态 %d", response.StatusCode)
	}
	return nil
}

// loginParams 从闲鱼登录页提取生成二维码所需的隐藏表单字段。
func (m *QRLoginManager) loginParams(ctx context.Context, session *QRSession) (url.Values, error) {
	loginURL, err := url.Parse(m.endpoints.miniLogin)
	if err != nil {
		return nil, err
	}
	query := loginURL.Query()
	for key, value := range map[string]string{
		"lang": "zh_cn", "appName": "xianyu", "appEntrance": "web", "styleType": "vertical",
		"bizParams": "", "notLoadSsoView": "false", "notKeepLogin": "false", "isMobile": "false",
		"qrCodeFirst": "false", "stie": "77", "rnd": fmt.Sprintf("%d", time.Now().UnixNano()),
	} {
		query.Set(key, value)
	}
	loginURL.RawQuery = query.Encode()
	request, err := NewQRRequest(ctx, http.MethodGet, loginURL.String(), nil)
	if err != nil {
		return nil, err
	}
	response, err := session.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("登录页 HTTP 状态 %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	viewData, err := extractViewData(body)
	if err != nil {
		return nil, err
	}
	params := make(url.Values, len(viewData.LoginFormData)+1)
	for key, value := range viewData.LoginFormData {
		params.Set(key, loginParam(value))
	}
	params.Set("umidTag", "SERVER")
	return params, nil
}

// loginParam 将登录页字段稳定编码为表单文本。
func loginParam(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(encoded)
	}
}

// pollOnce 查询一次二维码状态，并返回登录令牌或安全验证地址。
func (m *QRLoginManager) pollOnce(ctx context.Context, session *QRSession) (string, string, string, bool, error) {
	request, err := NewQRRequest(ctx, http.MethodPost, m.endpoints.query, strings.NewReader(session.params.Encode()))
	if err != nil {
		return "", "", "", false, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var payload struct {
		Content struct {
			Data struct {
				Status            string `json:"qrCodeStatus"`      // Status 是闲鱼二维码状态。
				Token             string `json:"token"`             // Token 是主要登录交换令牌。
				LGToken           string `json:"lgToken"`           // LGToken 是兼容旧响应的登录令牌。
				IframeRedirect    bool   `json:"iframeRedirect"`    // IframeRedirect 表示需要额外验证。
				IframeRedirectURL string `json:"iframeRedirectUrl"` // IframeRedirectURL 是验证地址。
			} `json:"data"` // Data 包含轮询状态与令牌。
		} `json:"content"` // Content 包含二维码业务结果。
	}
	if err := doJSON(session.client, request, &payload); err != nil {
		return "", "", "", false, err
	}
	loginToken := payload.Content.Data.Token
	if loginToken == "" {
		loginToken = payload.Content.Data.LGToken
	}
	if payload.Content.Data.IframeRedirect {
		return payload.Content.Data.Status, loginToken, safeVerificationURL(payload.Content.Data.IframeRedirectURL), true, nil
	}
	return payload.Content.Data.Status, loginToken, "", false, nil
}

// completeLogin 将手机确认后的 loginToken 交换为登录 Cookie。
func (m *QRLoginManager) completeLogin(ctx context.Context, session *QRSession, loginToken string) error {
	target, err := url.Parse(m.endpoints.loginToken)
	if err != nil {
		return err
	}
	query := target.Query()
	query.Set("token", loginToken)
	query.Set("subFlow", "DIALOG_CHECK_LOGIN_RPC")
	query.Set("nextCode", "0018")
	query.Set("bizScene", "qrcode")
	query.Set("confirm", "true")
	target.RawQuery = query.Encode()
	body := url.Values{"deviceId": {cookieValue(session.client, m.endpoints.h5Token, "cna")}}
	request, err := NewQRRequest(ctx, http.MethodPost, target.String(), strings.NewReader(body.Encode()))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := session.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("登录交换 HTTP 状态 %d", response.StatusCode)
	}
	return nil
}

// safeVerificationURL 仅允许 HTTPS 的 goofish.com 安全验证地址。
func safeVerificationURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || (parsed.Hostname() != "goofish.com" && !strings.HasSuffix(parsed.Hostname(), ".goofish.com")) {
		return ""
	}
	return parsed.String()
}

// LoginViewData 是闲鱼登录页内嵌的二维码表单配置。
type LoginViewData struct {
	LoginFormData map[string]any `json:"loginFormData"` // LoginFormData 保存生成二维码的隐藏字段。
}

// extractViewData 从登录页脚本中解码非固定前缀的 viewData。
func extractViewData(body []byte) (LoginViewData, error) {
	text := string(body)
	for _, marker := range []string{"window.viewData", "var viewData"} {
		index := strings.Index(text, marker)
		if index < 0 {
			continue
		}
		rest := text[index+len(marker):]
		_, rest, ok := strings.Cut(rest, "=")
		if !ok {
			continue
		}
		var result LoginViewData
		if err := json.NewDecoder(strings.NewReader(strings.TrimSpace(rest))).Decode(&result); err == nil && len(result.LoginFormData) > 0 {
			return result, nil
		}
	}
	return LoginViewData{}, errors.New("登录页缺少 loginFormData")
}

// NewQRRequest 创建带闲鱼 Passport 浏览器请求头的 HTTP 请求。
func NewQRRequest(ctx context.Context, method, target string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, target, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120.0.0.0 Safari/537.36")
	request.Header.Set("Accept", "application/json, text/plain, */*")
	request.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	request.Header.Set("Referer", "https://passport.goofish.com/")
	request.Header.Set("Origin", "https://passport.goofish.com")
	return request, nil
}

// doJSON 限制响应大小、校验 HTTP 状态并解码 JSON。
func doJSON(client *http.Client, request *http.Request, target any) error {
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("闲鱼登录接口 HTTP 状态 %d", response.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("闲鱼登录接口响应格式无法识别（Content-Type: %s）: %w", response.Header.Get("Content-Type"), err)
	}
	return nil
}

// cookieValue 从 client 的 Cookie Jar 读取 target 域下的指定 Cookie。
func cookieValue(client *http.Client, target, name string) string {
	parsed, err := url.Parse(target)
	if err != nil {
		return ""
	}
	for _, cookie := range client.Jar.Cookies(parsed) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

// cookieHeader 合并多个目标域的 Cookie 并稳定排序为请求头文本。
func cookieHeader(client *http.Client, targets ...string) string {
	values := make(map[string]string)
	for _, target := range targets {
		parsed, err := url.Parse(target)
		if err != nil {
			continue
		}
		for _, cookie := range client.Jar.Cookies(parsed) {
			values[cookie.Name] = cookie.Value
		}
	}
	return mergeResponseCookies(values, nil)
}
