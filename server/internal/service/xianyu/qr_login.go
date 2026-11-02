package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model/xianyu"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/xianyu/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/xianyu/response"
	"github.com/google/uuid"
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
