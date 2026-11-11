package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// webSocketTokenEndpoint 是闲鱼消息令牌的 MTop 入口。
	webSocketTokenEndpoint = "https://h5api.m.goofish.com/h5/mtop.taobao.idlemessage.pc.login.token/1.0/"
	// deviceCookieEndpoint 是获取 cna 设备 Cookie 的入口。
	deviceCookieEndpoint = "https://log.mmstat.com/eg.js"
	// webSocketAppKey 是闲鱼消息协议应用标识。
	webSocketAppKey = "444e9908a51d1cb236a27862abc769c9"
	// mtopAppKey 是闲鱼 Web MTop 应用标识。
	mtopAppKey = "34839810"
	// webSocketTokenLifetime 是本地缓存消息令牌的保守有效期。
	webSocketTokenLifetime = 20 * time.Hour
)

// VerificationRequiredError 表示闲鱼要求用户完成网页安全验证。
type VerificationRequiredError struct {
	URL string // URL 是经过域名校验的验证地址。
}

// Error 返回不含敏感信息的用户提示。
func (e *VerificationRequiredError) Error() string {
	return "消息接口需要网页安全验证并更新 Cookie"
}

// NonRetryable 表明自动重连无法消除安全验证要求。
func (e *VerificationRequiredError) NonRetryable() bool { return true }

// ActionURL 返回用户需要访问的安全验证地址。
func (e *VerificationRequiredError) ActionURL() string { return e.URL }

// CredentialExpiredError 表示闲鱼登录凭据已失效。
type CredentialExpiredError struct{}

// Error 返回重新登录提示。
func (CredentialExpiredError) Error() string { return "闲鱼登录已过期，请重新扫码登录" }

// NonRetryable 表明自动重连无法恢复已失效的登录凭据。
func (CredentialExpiredError) NonRetryable() bool { return true }

// WebSocketToken 获取、缓存并加密保存闲鱼消息连接令牌。
type WebSocketToken struct {
	db                   *gorm.DB       // db 读取账号凭据并保存刷新结果。
	events               *BusinessEvent // events 记录安全验证等业务结果。
	client               *http.Client   // client 是可在测试中替换的 HTTP 客户端。
	endpoint             string         // endpoint 是可在测试中替换的令牌入口。
	deviceCookieEndpoint string         // deviceCookieEndpoint 是可替换的设备 Cookie 入口。
}

// NewWebSocketToken 创建使用官方端点的消息令牌服务。
func NewWebSocketToken(db *gorm.DB, events *BusinessEvent) WebSocketToken {
	return WebSocketToken{
		db:                   db,
		events:               events,
		client:               &http.Client{Timeout: 30 * time.Second},
		endpoint:             webSocketTokenEndpoint,
		deviceCookieEndpoint: deviceCookieEndpoint,
	}
}

// Get 返回 ownerID 名下账号的有效令牌和设备标识，必要时刷新并原子保存。
func (s *WebSocketToken) Get(ctx context.Context, ownerID, accountID uint) (string, string, error) {
	if ownerID == 0 || accountID == 0 {
		return "", "", errors.New("账号 ID 不能为空")
	}
	lockValue, _ := credentialRefreshLocks.LoadOrStore(accountID, new(sync.Mutex))
	lock := lockValue.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	db := s.db.WithContext(ctx)
	var account model.Account
	if err := db.Where("id = ? AND owner_id = ?", accountID, ownerID).First(&account).Error; err != nil {
		return "", "", err
	}
	var credential model.Credential
	if err := db.Where("account_id = ?", accountID).First(&credential).Error; err != nil {
		return "", "", err
	}
	if credential.TokenCiphertext != "" && credential.TokenExpiresAt != nil && time.Until(*credential.TokenExpiresAt) > 5*time.Minute {
		token, err := decryptCredential(s.db, credential.TokenCiphertext, credentialAAD(accountID, "token"))
		return token, account.DeviceID, err
	}
	cookie, err := decryptCredential(s.db, credential.CookieCiphertext, credentialAAD(accountID, "cookie"))
	if err != nil {
		return "", "", err
	}
	cookie, err = s.ensureDeviceCookie(ctx, cookie)
	if err != nil {
		return "", "", err
	}
	cookies := parseCookieHeader(cookie)
	unb := cookies["unb"]
	if unb == "" {
		return "", "", errors.New("闲鱼凭证缺少用户标识")
	}
	deviceID := account.DeviceID
	if deviceID == "" {
		deviceID = strings.ToUpper(uuid.NewString()) + "-" + unb
	}

	initialCookie := cookie
	var token string
	for attempt := range 2 {
		token, cookie, err = s.fetch(ctx, cookie, deviceID)
		if err == nil {
			break
		}
		if verification, ok := errors.AsType[*VerificationRequiredError](err); ok {
			cipherCookie, encryptErr := encryptCredential(s.db, cookie, credentialAAD(accountID, "cookie"))
			if encryptErr != nil {
				return "", "", errors.Join(verification, fmt.Errorf("保存风控 Cookie: %w", encryptErr))
			}
			statusErr := db.Transaction(func(tx *gorm.DB) error {
				result := tx.Model(&model.Credential{}).
					Where("id = ? AND account_id = ? AND version = ?", credential.ID, accountID, credential.Version).
					Updates(map[string]any{"cookie_ciphertext": cipherCookie, "version": gorm.Expr("version + 1")})
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected != 1 {
					return errors.New("凭证已被其他操作更新，请重试")
				}
				return tx.Model(&model.Account{}).Where("id = ? AND owner_id = ?", accountID, ownerID).
					Updates(map[string]any{"device_id": deviceID, "status": "verification_required"}).Error
			})
			s.events.Record(ctx, model.BusinessEvent{
				OwnerID: ownerID, AccountID: accountID, Type: "websocket", Status: "failed",
				Summary: "闲鱼消息连接需要安全验证", Error: verification.Error(),
			})
			if statusErr != nil {
				return "", "", errors.Join(verification, fmt.Errorf("更新账号验证状态: %w", statusErr))
			}
			return "", "", verification
		}
		if _, expired := errors.AsType[CredentialExpiredError](err); expired {
			var cipherCookie string
			if cookie != initialCookie {
				cipherCookie, err = encryptCredential(s.db, cookie, credentialAAD(accountID, "cookie"))
				if err != nil {
					return "", "", errors.Join(CredentialExpiredError{}, fmt.Errorf("保存闲鱼更新后的 Cookie: %w", err))
				}
			}
			statusErr := db.Transaction(func(tx *gorm.DB) error {
				if cipherCookie != "" {
					result := tx.Model(&model.Credential{}).
						Where("id = ? AND account_id = ? AND version = ?", credential.ID, accountID, credential.Version).
						Updates(map[string]any{"cookie_ciphertext": cipherCookie, "version": gorm.Expr("version + 1")})
					if result.Error != nil {
						return result.Error
					}
					if result.RowsAffected != 1 {
						return errors.New("凭证已被其他操作更新，请重试")
					}
				}
				return tx.Model(&model.Account{}).Where("id = ? AND owner_id = ?", accountID, ownerID).Update("status", "expired").Error
			})
			return "", "", errors.Join(CredentialExpiredError{}, statusErr)
		}
		if attempt == 1 || !waitContext(ctx, 500*time.Millisecond) {
			if cookie != initialCookie {
				cipherCookie, encryptErr := encryptCredential(s.db, cookie, credentialAAD(accountID, "cookie"))
				if encryptErr != nil {
					return "", "", errors.Join(err, fmt.Errorf("保存闲鱼更新后的 Cookie: %w", encryptErr))
				}
				persistResult := db.Model(&model.Credential{}).
					Where("id = ? AND account_id = ? AND version = ?", credential.ID, accountID, credential.Version).
					Updates(map[string]any{"cookie_ciphertext": cipherCookie, "version": gorm.Expr("version + 1")})
				if persistResult.Error != nil {
					return "", "", errors.Join(err, fmt.Errorf("保存闲鱼更新后的 Cookie: %w", persistResult.Error))
				}
				if persistResult.RowsAffected != 1 {
					return "", "", errors.Join(err, errors.New("保存闲鱼更新后的 Cookie: 凭证已被其他操作更新，请重试"))
				}
			}
			return "", "", err
		}
	}
	expiresAt := time.Now().Add(webSocketTokenLifetime)
	cipherCookie, err := encryptCredential(s.db, cookie, credentialAAD(accountID, "cookie"))
	if err != nil {
		return "", "", err
	}
	cipherToken, err := encryptCredential(s.db, token, credentialAAD(accountID, "token"))
	if err != nil {
		return "", "", err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.Credential{}).
			Where("id = ? AND account_id = ? AND version = ?", credential.ID, accountID, credential.Version).
			Updates(map[string]any{
				"cookie_ciphertext": cipherCookie,
				"token_ciphertext":  cipherToken,
				"token_expires_at":  expiresAt,
				"version":           gorm.Expr("version + 1"),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("凭证已被其他操作更新，请重试")
		}
		return tx.Model(&model.Account{}).
			Where("id = ? AND owner_id = ?", accountID, ownerID).
			Updates(map[string]any{"device_id": deviceID, "status": "ready"}).Error
	})
	return token, deviceID, err
}

// ensureDeviceCookie 在原 Cookie 缺少 cna 时从统计入口获取并合并。
func (s *WebSocketToken) ensureDeviceCookie(ctx context.Context, cookie string) (string, error) {
	if parseCookieHeader(cookie)["cna"] != "" {
		return cookie, nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.deviceCookieEndpoint, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
	response, err := s.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("获取闲鱼设备 Cookie: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<20))
	for _, responseCookie := range response.Cookies() {
		if responseCookie.Name == "cna" && responseCookie.Value != "" {
			return mergeResponseCookies(parseCookieHeader(cookie), []*http.Cookie{responseCookie}), nil
		}
	}
	return "", errors.New("闲鱼设备 Cookie 获取失败")
}

// fetch 签名调用消息令牌接口，并返回令牌与响应合并后的 Cookie。
func (s *WebSocketToken) fetch(ctx context.Context, cookie, deviceID string) (string, string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	data, err := json.Marshal(struct {
		AppKey   string `json:"appKey"`   // AppKey 是消息协议应用标识。
		DeviceID string `json:"deviceId"` // DeviceID 是当前账号设备标识。
	}{AppKey: webSocketAppKey, DeviceID: deviceID})
	if err != nil {
		return "", cookie, err
	}
	mh5Token := strings.Split(parseCookieHeader(cookie)["_m_h5_tk"], "_")[0]
	hash := md5.Sum([]byte(mh5Token + "&" + timestamp + "&" + mtopAppKey + "&" + string(data)))
	target, err := url.Parse(s.endpoint)
	if err != nil {
		return "", cookie, err
	}
	query := target.Query()
	for key, value := range map[string]string{
		"jsv": "2.7.2", "appKey": mtopAppKey, "t": timestamp,
		"sign": hex.EncodeToString(hash[:]), "v": "1.0", "type": "originaljson",
		"accountSite": "xianyu", "dataType": "json", "timeout": "20000",
		"api": "mtop.taobao.idlemessage.pc.login.token", "sessionOption": "AutoLoginOnly",
		"dangerouslySetWindvaneParams": "%5Bobject%20Object%5D", "smToken": "token", "queryToken": "sm", "sm": "sm",
		"spm_cnt": "a21ybx.im.0.0", "spm_pre": "a21ybx.home.sidebar.1.4c053da6vYwnmf", "log_id": "4c053da6vYwnmf",
	} {
		query.Set(key, value)
	}
	target.RawQuery = query.Encode()
	form := url.Values{"data": {string(data)}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return "", cookie, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Origin", "https://www.goofish.com")
	request.Header.Set("Referer", "https://www.goofish.com/")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36")
	request.Header.Set("Sec-CH-UA", `"Chromium";v="146", "Not-A.Brand";v="24", "Google Chrome";v="146"`)
	request.Header.Set("Sec-CH-UA-Mobile", "?0")
	request.Header.Set("Sec-CH-UA-Platform", `"Windows"`)
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Site", "same-site")
	request.Header.Set("Accept-Language", "en,zh-CN;q=0.9,zh;q=0.8,zh-TW;q=0.7,ja;q=0.6")
	request.Header.Set("Priority", "u=1, i")
	request.Header.Set("Cookie", cookie)
	response, err := s.client.Do(request)
	if err != nil {
		return "", cookie, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", cookie, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", cookie, fmt.Errorf("闲鱼 Token 接口 HTTP 状态 %d", response.StatusCode)
	}
	mergedCookie := mergeResponseCookies(parseCookieHeader(cookie), response.Cookies())
	var payload struct {
		Ret  json.RawMessage `json:"ret"` // Ret 包含 MTop 状态码与消息。
		Data struct {
			AccessToken string `json:"accessToken"` // AccessToken 是消息连接令牌。
			URL         string `json:"url"`         // URL 是安全验证地址。
		} `json:"data"` // Data 包含令牌或验证信息。
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", mergedCookie, errors.New("闲鱼 Token 响应格式无法识别")
	}
	var ret []string
	if err := json.Unmarshal(payload.Ret, &ret); err != nil {
		var singleRet string
		if json.Unmarshal(payload.Ret, &singleRet) == nil {
			ret = []string{singleRet}
		}
	}
	for _, result := range ret {
		if strings.Contains(result, "FAIL_SYS_SESSION_EXPIRED") || strings.Contains(result, "FAIL_SYS_TOKEN_EXOIRED") || strings.Contains(result, "FAIL_SYS_TOKEN_EXPIRED") {
			return "", mergedCookie, CredentialExpiredError{}
		}
		if strings.Contains(result, "FAIL_SYS_USER_VALIDATE") || strings.Contains(result, "RGV587_ERROR") || strings.Contains(result, "被挤爆啦") {
			return "", mergedCookie, &VerificationRequiredError{URL: safeVerificationURL(payload.Data.URL)}
		}
	}
	for _, result := range ret {
		if strings.Contains(result, "SUCCESS::调用成功") {
			if payload.Data.AccessToken != "" {
				return payload.Data.AccessToken, mergedCookie, nil
			}
			return "", mergedCookie, errors.New("闲鱼 Token 接口返回成功但缺少 WebSocket Token")
		}
	}
	if len(ret) > 0 {
		code, _, _ := strings.Cut(ret[0], "::")
		return "", mergedCookie, fmt.Errorf("闲鱼 Token 接口失败（%s）", code)
	}
	if payload.Data.AccessToken == "" {
		return "", mergedCookie, errors.New("闲鱼未返回 WebSocket Token")
	}
	return "", mergedCookie, errors.New("闲鱼 Token 接口未确认请求成功")
}
