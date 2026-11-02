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

	"github.com/gcc798/go.xianyu.assistant/server/internal/infra"
	"github.com/gcc798/go.xianyu.assistant/server/internal/model/xianyu"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	webSocketTokenEndpoint = "https://h5api.m.goofish.com/h5/mtop.taobao.idlemessage.pc.login.token/1.0/"
	deviceCookieEndpoint   = "https://log.mmstat.com/eg.js"
	webSocketAppKey        = "444e9908a51d1cb236a27862abc769c9"
	mtopAppKey             = "34839810"
	webSocketTokenLifetime = 20 * time.Hour
)

type verificationRequiredError struct {
	URL string
}

func (e *verificationRequiredError) Error() string {
	return "消息接口需要网页安全验证并更新 Cookie"
}
func (e *verificationRequiredError) NonRetryable() bool { return true }
func (e *verificationRequiredError) ActionURL() string  { return e.URL }

type webSocketToken struct {
	client               *http.Client
	endpoint             string
	deviceCookieEndpoint string
}

func newWebSocketToken() webSocketToken {
	return webSocketToken{
		client:               &http.Client{Timeout: 30 * time.Second},
		endpoint:             webSocketTokenEndpoint,
		deviceCookieEndpoint: deviceCookieEndpoint,
	}
}

func (s *webSocketToken) Get(ctx context.Context, ownerID, accountID uint) (string, string, error) {
	if ownerID == 0 || accountID == 0 {
		return "", "", errors.New("账号 ID 不能为空")
	}
	lockValue, _ := credentialRefreshLocks.LoadOrStore(accountID, new(sync.Mutex))
	lock := lockValue.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	db := infra.DB.WithContext(ctx)
	var account model.Account
	if err := db.Where("id = ? AND owner_id = ?", accountID, ownerID).First(&account).Error; err != nil {
		return "", "", err
	}
	var credential model.Credential
	if err := db.Where("account_id = ?", accountID).First(&credential).Error; err != nil {
		return "", "", err
	}
	if credential.TokenCiphertext != "" && credential.TokenExpiresAt != nil && time.Until(*credential.TokenExpiresAt) > 5*time.Minute {
		token, err := decryptCredential(credential.TokenCiphertext, credentialAAD(accountID, "token"))
		return token, account.DeviceID, err
	}
	cookie, err := decryptCredential(credential.CookieCiphertext, credentialAAD(accountID, "cookie"))
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

	var token string
	for attempt := range 2 {
		token, cookie, err = s.fetch(ctx, cookie, deviceID)
		if err == nil {
			break
		}
		if verification, ok := errors.AsType[*verificationRequiredError](err); ok {
			cipherCookie, encryptErr := encryptCredential(cookie, credentialAAD(accountID, "cookie"))
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
			recordBusinessEvent(ctx, model.BusinessEvent{
				OwnerID: ownerID, AccountID: accountID, Type: "websocket", Status: "failed",
				Summary: "闲鱼消息连接需要安全验证", Error: verification.Error(),
			})
			if statusErr != nil {
				return "", "", errors.Join(verification, fmt.Errorf("更新账号验证状态: %w", statusErr))
			}
			return "", "", verification
		}
		if attempt == 1 || !waitContext(ctx, 500*time.Millisecond) {
			return "", "", err
		}
	}
	expiresAt := time.Now().Add(webSocketTokenLifetime)
	cipherCookie, err := encryptCredential(cookie, credentialAAD(accountID, "cookie"))
	if err != nil {
		return "", "", err
	}
	cipherToken, err := encryptCredential(token, credentialAAD(accountID, "token"))
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

func (s *webSocketToken) ensureDeviceCookie(ctx context.Context, cookie string) (string, error) {
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

func (s *webSocketToken) fetch(ctx context.Context, cookie, deviceID string) (string, string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	data, err := json.Marshal(struct {
		AppKey   string `json:"appKey"`
		DeviceID string `json:"deviceId"`
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
		"spm_cnt": "a21ybx.im.0.0", "spm_pre": "a21ybx.item.want.1.14ad3da6ALVq3n", "log_id": "14ad3da6ALVq3n",
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
		Ret  []string `json:"ret"`
		Data struct {
			AccessToken string `json:"accessToken"`
			URL         string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", mergedCookie, errors.New("闲鱼 Token 响应格式无法识别")
	}
	for _, result := range payload.Ret {
		if strings.Contains(result, "FAIL_SYS_USER_VALIDATE") || strings.Contains(result, "RGV587_ERROR") || strings.Contains(result, "被挤爆啦") {
			return "", mergedCookie, &verificationRequiredError{URL: safeVerificationURL(payload.Data.URL)}
		}
	}
	if payload.Data.AccessToken == "" {
		return "", mergedCookie, errors.New("闲鱼未返回 WebSocket Token")
	}
	return payload.Data.AccessToken, mergedCookie, nil
}
