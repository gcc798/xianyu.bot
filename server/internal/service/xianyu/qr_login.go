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
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model/xianyu"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/xianyu/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/xianyu/response"
	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
)

const (
	qrStatusPending              = "pending"
	qrStatusScanned              = "scanned"
	qrStatusConfirmed            = "confirmed"
	qrStatusExpired              = "expired"
	qrStatusCancelled            = "cancelled"
	qrStatusVerificationRequired = "verification_required"
	qrStatusError                = "error"
)

type qrEndpoints struct {
	h5Token    string
	miniLogin  string
	generate   string
	query      string
	loginToken string
}

type qrLogin struct {
	manager *qrLoginManager
}

type qrLoginManager struct {
	mu        sync.Mutex
	root      context.Context
	accounts  *account
	sessions  map[string]*qrSession
	byAccount map[[2]uint]string
	endpoints qrEndpoints
	timeout   time.Duration
	pollEvery time.Duration
}

type qrSession struct {
	mu              sync.Mutex
	id              string
	ownerID         uint
	accountID       uint
	status          string
	message         string
	qrCodeDataURL   string
	verificationURL string
	params          url.Values
	client          *http.Client
	createdAt       time.Time
	expiresAt       time.Time
	cancel          context.CancelFunc
}

type stringOrNumber string

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

func newQRLogin(root context.Context, accounts *account) qrLogin {
	return qrLogin{manager: &qrLoginManager{
		root:      contextOrBackground(root),
		accounts:  accounts,
		sessions:  make(map[string]*qrSession),
		byAccount: make(map[[2]uint]string),
		endpoints: qrEndpoints{
			h5Token:    "https://h5api.m.goofish.com/h5/mtop.gaia.nodejs.gaia.idle.data.gw.v2.index.get/1.0/",
			miniLogin:  "https://passport.goofish.com/mini_login.htm",
			generate:   "https://passport.goofish.com/newlogin/qrcode/generate.do",
			query:      "https://passport.goofish.com/newlogin/qrcode/query.do",
			loginToken: "https://passport.goofish.com/login_token/login.do",
		},
		timeout:   5 * time.Minute,
		pollEvery: time.Second,
	}}
}

func (s *qrLogin) Start(ctx context.Context, ownerID, accountID uint) (xyResponse.QRLogin, error) {
	if ownerID == 0 || accountID == 0 {
		return xyResponse.QRLogin{}, errors.New("账号 ID 不能为空")
	}
	var account model.Account
	if err := s.manager.accounts.db.WithContext(ctx).Where("id = ? AND owner_id = ?", accountID, ownerID).First(&account).Error; err != nil {
		return xyResponse.QRLogin{}, err
	}
	manager := s.manager
	if manager == nil {
		return xyResponse.QRLogin{}, errors.New("二维码登录服务未初始化")
	}
	if existing := manager.active(ownerID, accountID); existing != nil {
		return existing.response(true), nil
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return xyResponse.QRLogin{}, err
	}
	sessionCtx, cancel := context.WithTimeout(contextOrBackground(manager.root), manager.timeout)
	session := &qrSession{
		id:        uuid.NewString(),
		ownerID:   ownerID,
		accountID: accountID,
		status:    qrStatusPending,
		message:   "等待扫码",
		params:    make(url.Values),
		client:    &http.Client{Jar: jar, Timeout: 30 * time.Second},
		createdAt: time.Now(),
		expiresAt: time.Now().Add(manager.timeout),
		cancel:    cancel,
	}
	if err := manager.prepare(sessionCtx, session); err != nil {
		cancel()
		return xyResponse.QRLogin{}, err
	}
	manager.store(session)
	go manager.poll(sessionCtx, session)
	return session.response(true), nil
}

func (s *qrLogin) Status(ownerID uint, sessionID string) (xyResponse.QRLogin, error) {
	session := s.manager.get(ownerID, sessionID)
	if session == nil {
		return xyResponse.QRLogin{}, errors.New("二维码会话不存在或已过期")
	}
	return session.response(false), nil
}

func (s *qrLogin) Cancel(ownerID uint, sessionID string) error {
	session := s.manager.get(ownerID, sessionID)
	if session == nil {
		return errors.New("二维码会话不存在或已过期")
	}
	session.finish(qrStatusCancelled, "登录已取消", "")
	return nil
}

func (m *qrLoginManager) active(ownerID, accountID uint) *qrSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupLocked(time.Now())
	key := [2]uint{ownerID, accountID}
	session := m.sessions[m.byAccount[key]]
	if session == nil || session.terminal() || time.Now().After(session.expiresAt) {
		delete(m.byAccount, key)
		return nil
	}
	return session
}

func (m *qrLoginManager) cleanupLocked(now time.Time) {
	for id, session := range m.sessions {
		if now.Before(session.expiresAt.Add(5 * time.Minute)) {
			continue
		}
		delete(m.sessions, id)
		key := [2]uint{session.ownerID, session.accountID}
		if m.byAccount[key] == id {
			delete(m.byAccount, key)
		}
	}
}

func (m *qrLoginManager) store(session *qrSession) {
	m.mu.Lock()
	m.sessions[session.id] = session
	m.byAccount[[2]uint{session.ownerID, session.accountID}] = session.id
	m.mu.Unlock()
}

func (m *qrLoginManager) get(ownerID uint, sessionID string) *qrSession {
	m.mu.Lock()
	session := m.sessions[sessionID]
	m.mu.Unlock()
	if session == nil || session.ownerID != ownerID {
		return nil
	}
	return session
}

func (m *qrLoginManager) prepare(ctx context.Context, session *qrSession) error {
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
	request, err := newQRRequest(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return err
	}
	var payload struct {
		Content struct {
			Success bool `json:"success"`
			Data    struct {
				T           stringOrNumber `json:"t"`
				CK          stringOrNumber `json:"ck"`
				CodeContent string         `json:"codeContent"`
			} `json:"data"`
		} `json:"content"`
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

func (m *qrLoginManager) refreshH5Token(ctx context.Context, session *qrSession) error {
	request, err := newQRRequest(ctx, http.MethodGet, m.endpoints.h5Token, nil)
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
	request, err = newQRRequest(ctx, http.MethodPost, signedURL.String(), strings.NewReader(""))
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

func (m *qrLoginManager) loginParams(ctx context.Context, session *qrSession) (url.Values, error) {
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
	request, err := newQRRequest(ctx, http.MethodGet, loginURL.String(), nil)
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

func (m *qrLoginManager) poll(ctx context.Context, session *qrSession) {
	ticker := time.NewTicker(m.pollEvery)
	defer ticker.Stop()
	defer func() {
		if errors.Is(context.Cause(ctx), context.DeadlineExceeded) {
			session.finish(qrStatusExpired, "二维码已过期", "")
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			status, loginToken, verificationURL, verificationRequired, err := m.pollOnce(ctx, session)
			if err != nil {
				continue
			}
			switch status {
			case "NEW":
			case "SCANED":
				session.setStatus(qrStatusScanned, "已扫码，等待手机确认")
			case "CONFIRMED":
				if verificationRequired {
					session.finish(qrStatusVerificationRequired, "需要完成闲鱼安全验证", verificationURL)
					return
				}
				if loginToken != "" {
					if err := m.completeLogin(ctx, session, loginToken); err != nil {
						session.finish(qrStatusError, "扫码确认后的登录交换失败", "")
						return
					}
				}
				if err := m.refreshH5Token(ctx, session); err != nil {
					session.finish(qrStatusError, "登录后的闲鱼凭证刷新失败", "")
					return
				}
				if err := m.saveCredential(ctx, session); err != nil {
					session.finish(qrStatusError, "登录凭证保存失败", "")
					return
				}
				session.finish(qrStatusConfirmed, "登录成功，凭证已加密保存", "")
				return
			case "EXPIRED":
				session.finish(qrStatusExpired, "二维码已过期", "")
				return
			default:
				session.finish(qrStatusCancelled, "登录已取消", "")
				return
			}
		}
	}
}

func (m *qrLoginManager) pollOnce(ctx context.Context, session *qrSession) (string, string, string, bool, error) {
	request, err := newQRRequest(ctx, http.MethodPost, m.endpoints.query, strings.NewReader(session.params.Encode()))
	if err != nil {
		return "", "", "", false, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var payload struct {
		Content struct {
			Data struct {
				Status            string `json:"qrCodeStatus"`
				Token             string `json:"token"`
				LGToken           string `json:"lgToken"`
				IframeRedirect    bool   `json:"iframeRedirect"`
				IframeRedirectURL string `json:"iframeRedirectUrl"`
			} `json:"data"`
		} `json:"content"`
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

func (m *qrLoginManager) completeLogin(ctx context.Context, session *qrSession, loginToken string) error {
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
	request, err := newQRRequest(ctx, http.MethodPost, target.String(), strings.NewReader(body.Encode()))
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

func safeVerificationURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || (parsed.Hostname() != "goofish.com" && !strings.HasSuffix(parsed.Hostname(), ".goofish.com")) {
		return ""
	}
	return parsed.String()
}

func (m *qrLoginManager) saveCredential(ctx context.Context, session *qrSession) error {
	cookie := cookieHeader(session.client, m.endpoints.h5Token, m.endpoints.miniLogin, m.endpoints.query)
	unb := parseCookieHeader(cookie)["unb"]
	if cookie == "" || unb == "" {
		return errors.New("扫码响应缺少账号 Cookie")
	}
	if err := m.accounts.SaveCredential(ctx, session.ownerID, xyRequest.SaveCredential{ID: session.accountID, Cookie: cookie}); err != nil {
		return err
	}
	return m.accounts.db.WithContext(ctx).Model(&model.Account{}).
		Where("id = ? AND owner_id = ?", session.accountID, session.ownerID).
		Update("external_user_id", unb).Error
}

type loginViewData struct {
	LoginFormData map[string]any `json:"loginFormData"`
}

func extractViewData(body []byte) (loginViewData, error) {
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
		var result loginViewData
		if err := json.NewDecoder(strings.NewReader(strings.TrimSpace(rest))).Decode(&result); err == nil && len(result.LoginFormData) > 0 {
			return result, nil
		}
	}
	return loginViewData{}, errors.New("登录页缺少 loginFormData")
}

func newQRRequest(ctx context.Context, method, target string, body io.Reader) (*http.Request, error) {
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

func (s *qrSession) response(includeQR bool) xyResponse.QRLogin {
	s.mu.Lock()
	defer s.mu.Unlock()
	response := xyResponse.QRLogin{
		SessionID: s.id, AccountID: s.accountID, Status: s.status,
		Message: s.message, VerificationURL: s.verificationURL,
	}
	if includeQR {
		response.QRCodeDataURL = s.qrCodeDataURL
	}
	return response
}

func (s *qrSession) terminal() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status == qrStatusConfirmed || s.status == qrStatusExpired || s.status == qrStatusCancelled || s.status == qrStatusVerificationRequired || s.status == qrStatusError
}

func (s *qrSession) setStatus(status, message string) {
	s.mu.Lock()
	if !s.isTerminalLocked() {
		s.status = status
		s.message = message
	}
	s.mu.Unlock()
}

func (s *qrSession) finish(status, message, verificationURL string) {
	s.mu.Lock()
	if !s.isTerminalLocked() {
		s.status = status
		s.message = message
		s.verificationURL = verificationURL
		s.cancel()
	}
	s.mu.Unlock()
}

func (s *qrSession) isTerminalLocked() bool {
	return s.status == qrStatusConfirmed || s.status == qrStatusExpired || s.status == qrStatusCancelled || s.status == qrStatusVerificationRequired || s.status == qrStatusError
}
