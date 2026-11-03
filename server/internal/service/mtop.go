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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type mtopCaller interface {
	Call(context.Context, uint, uint, string, any, bool) (json.RawMessage, error)
}

type MTopClient struct {
	db       *gorm.DB
	client   *http.Client
	endpoint string
}

type MTopEnvelope struct {
	Ret  []string        `json:"ret"`
	Data json.RawMessage `json:"data"`
}

var errMTopTransport = errors.New("闲鱼接口结果未知")

func (m *MTopClient) Call(ctx context.Context, ownerID, accountID uint, api string, data any, seller bool) (result json.RawMessage, err error) {
	startedAt := time.Now()
	defer func() {
		fields := []zap.Field{
			zap.Uint("ownerID", ownerID), zap.Uint("accountID", accountID), zap.String("api", api),
			zap.Bool("seller", seller), zap.Duration("duration", time.Since(startedAt)),
		}
		if err != nil {
			zap.L().Warn("调用闲鱼 MTop 接口失败", append(fields, zap.Error(err))...)
		} else {
			zap.L().Info("调用闲鱼 MTop 接口成功", fields...)
		}
	}()
	lockValue, _ := credentialRefreshLocks.LoadOrStore(accountID, new(sync.Mutex))
	lock := lockValue.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	db := m.db.WithContext(ctx)
	if _, err := ownedAccount(ctx, m.db, ownerID, accountID); err != nil {
		return nil, err
	}
	var credential model.Credential
	if err := db.Where("account_id = ?", accountID).First(&credential).Error; err != nil {
		return nil, err
	}
	cookie, err := decryptCredential(m.db, credential.CookieCiphertext, credentialAAD(accountID, "cookie"))
	if err != nil {
		return nil, err
	}
	originalCookie := cookie
	defer func() {
		if cookie == originalCookie {
			return
		}
		ciphertext, saveErr := encryptCredential(m.db, cookie, credentialAAD(accountID, "cookie"))
		if saveErr == nil {
			update := db.Model(&model.Credential{}).Where("id = ? AND version = ?", credential.ID, credential.Version).
				Updates(map[string]any{"cookie_ciphertext": ciphertext, "version": gorm.Expr("version + 1")})
			saveErr = update.Error
			if saveErr == nil && update.RowsAffected != 1 {
				saveErr = errors.New("凭证已被其他操作更新，请重试")
			}
		}
		if err == nil && saveErr != nil {
			err = saveErr
		}
	}()

	for attempt := range 2 {
		var envelope MTopEnvelope
		var nextCookie string
		envelope, nextCookie, err = m.call(ctx, cookie, api, data, seller)
		cookie = nextCookie
		if err == nil {
			return envelope.Data, nil
		}
		if attempt == 0 && isMTopTokenError(err) && cookie != originalCookie {
			continue
		}
		return nil, err
	}
	return nil, err
}

func (m *MTopClient) call(ctx context.Context, cookie, api string, data any, seller bool) (MTopEnvelope, string, error) {
	version := "1.0"
	if strings.HasSuffix(api, "/2.0") {
		api = strings.TrimSuffix(api, "/2.0")
		version = "2.0"
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return MTopEnvelope{}, cookie, err
	}
	cookies := parseCookieHeader(cookie)
	token := strings.Split(cookies["_m_h5_tk"], "_")[0]
	if token == "" {
		return MTopEnvelope{}, cookie, errors.New("闲鱼凭证缺少签名令牌，请重新扫码登录")
	}
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	hash := md5.Sum([]byte(token + "&" + timestamp + "&" + mtopAppKey + "&" + string(payload)))
	base := strings.TrimRight(m.endpoint, "/")
	if base == "" {
		base = "https://h5api.m.goofish.com/h5"
	}
	target, err := url.Parse(base + "/" + api + "/" + version + "/")
	if err != nil {
		return MTopEnvelope{}, cookie, err
	}
	query := target.Query()
	for key, value := range map[string]string{
		"jsv": "2.7.2", "appKey": mtopAppKey, "t": timestamp, "sign": hex.EncodeToString(hash[:]),
		"v": version, "type": "originaljson", "dataType": "json", "accountSite": "xianyu",
		"api": api, "sessionOption": "AutoLoginOnly", "timeout": "20000",
	} {
		query.Set(key, value)
	}
	if seller {
		query.Set("type", "json")
		query.Set("valueType", "string")
		query.Set("idle_site_biz_code", "COMMONPRO")
		query.Set("spm_cnt", "a21107h.42826273.0.0")
	}
	if strings.Contains(api, ".detail") {
		query.Set("spm_cnt", "a21ybx.item.0.0")
	}
	target.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), strings.NewReader(url.Values{"data": {string(payload)}}.Encode()))
	if err != nil {
		return MTopEnvelope{}, cookie, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", webBrowserUA)
	request.Header.Set("Cookie", cookie)
	if seller {
		request.Header.Set("idle_site_biz_code", "COMMONPRO")
		request.Header.Set("Origin", "https://seller.goofish.com")
		request.Header.Set("Referer", "https://seller.goofish.com/?site=COMMONPRO")
	} else {
		request.Header.Set("Origin", "https://www.goofish.com")
		referer := "https://www.goofish.com/"
		if strings.Contains(api, ".detail") {
			if values, ok := data.(map[string]any); ok && fmt.Sprint(values["itemId"]) != "" {
				referer = "https://www.goofish.com/item?id=" + url.QueryEscape(fmt.Sprint(values["itemId"]))
			}
		}
		request.Header.Set("Referer", referer)
	}
	client := m.client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return MTopEnvelope{}, cookie, fmt.Errorf("%w: %v", errMTopTransport, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 8<<20))
	merged := mergeResponseCookies(cookies, response.Cookies())
	if err != nil {
		return MTopEnvelope{}, merged, fmt.Errorf("%w: %v", errMTopTransport, err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return MTopEnvelope{}, merged, fmt.Errorf("闲鱼接口 HTTP 状态 %d", response.StatusCode)
	}
	var envelope MTopEnvelope
	if json.Unmarshal(body, &envelope) != nil || len(envelope.Ret) == 0 {
		return MTopEnvelope{}, merged, errors.New("闲鱼接口响应格式无法识别")
	}
	if !strings.Contains(envelope.Ret[0], "SUCCESS") {
		return MTopEnvelope{}, merged, errors.New(envelope.Ret[0])
	}
	return envelope, merged, nil
}

func isMTopTokenError(err error) bool {
	message := err.Error()
	return strings.Contains(message, "TOKEN_EXPIRED") || strings.Contains(message, "TOKEN_EXOIRED") || strings.Contains(message, "TOKEN_EMPTY")
}
