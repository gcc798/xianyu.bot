package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
	"time"

	"github.com/gcc798/xianyu.bot/internal/model"
	xyRequest "github.com/gcc798/xianyu.bot/internal/model/request"
	xyResponse "github.com/gcc798/xianyu.bot/internal/model/response"
	"github.com/google/uuid"
)

const (
	// qrStatusPending 表示等待用户扫码。
	qrStatusPending = "pending"
	// qrStatusScanned 表示已扫码且等待手机确认。
	qrStatusScanned = "scanned"
	// qrStatusConfirmed 表示登录成功且凭据已保存。
	qrStatusConfirmed = "confirmed"
	// qrStatusExpired 表示二维码会话超时。
	qrStatusExpired = "expired"
	// qrStatusCancelled 表示用户或远端取消登录。
	qrStatusCancelled = "cancelled"
	// qrStatusVerificationRequired 表示登录需要额外安全验证。
	qrStatusVerificationRequired = "verification_required"
	// qrStatusError 表示登录流程发生不可恢复错误。
	qrStatusError = "error"
)

// QREndpoints 汇总二维码登录流程使用的闲鱼入口。
type QREndpoints struct {
	h5Token    string // h5Token 初始化和刷新 H5 Cookie。
	miniLogin  string // miniLogin 返回二维码登录表单参数。
	generate   string // generate 生成二维码内容。
	query      string // query 轮询二维码状态。
	loginToken string // loginToken 将确认令牌交换为登录 Cookie。
}

// QRLogin 提供用户维度的二维码登录入口。
type QRLogin struct {
	manager *QRLoginManager // manager 管理并发登录会话与轮询。
}

// NewQRLogin 创建二维码登录服务及默认闲鱼端点。
func NewQRLogin(root context.Context, accounts *Account) *QRLogin {
	return &QRLogin{manager: &QRLoginManager{
		root:      contextOrBackground(root),
		accounts:  accounts,
		sessions:  make(map[string]*QRSession),
		byAccount: make(map[[2]uint]string),
		endpoints: QREndpoints{
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

// QRLoginManager 管理二维码会话索引、生命周期和远端轮询。
type QRLoginManager struct {
	mu        sync.Mutex            // mu 保护 sessions 和 byAccount。
	root      context.Context       // root 控制全部登录会话生命周期。
	accounts  *Account              // accounts 校验归属并保存加密凭据。
	sessions  map[string]*QRSession // sessions 按会话 ID 保存状态。
	byAccount map[[2]uint]string    // byAccount 限制每个用户账号只有一个活动会话。
	endpoints QREndpoints           // endpoints 是可在测试中替换的闲鱼入口。
	timeout   time.Duration         // timeout 是单个二维码会话有效期。
	pollEvery time.Duration         // pollEvery 是远端状态轮询间隔。
}

// QRSession 保存一次二维码登录的可变状态和独立 Cookie Jar。
type QRSession struct {
	mu              sync.Mutex         // mu 保护会话可变字段。
	id              string             // id 是对外会话标识。
	ownerID         uint               // ownerID 是会话所属系统用户 ID。
	accountID       uint               // accountID 是待登录的闲鱼账号 ID。
	status          string             // status 是当前登录状态。
	message         string             // message 是当前用户提示。
	qrCodeDataURL   string             // qrCodeDataURL 是首次响应展示的二维码图片。
	verificationURL string             // verificationURL 是额外安全验证地址。
	params          url.Values         // params 是生成与轮询二维码所需表单参数。
	client          *http.Client       // client 持有本会话独立 Cookie Jar。
	createdAt       time.Time          // createdAt 是会话创建时间。
	expiresAt       time.Time          // expiresAt 是会话到期时间。
	cancel          context.CancelFunc // cancel 终止轮询上下文。
}

// Start 校验账号归属并创建或复用活动二维码登录会话。
func (s *QRLogin) Start(ctx context.Context, ownerID, accountID uint) (xyResponse.QRLogin, error) {
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
	session := &QRSession{
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

// Status 返回 ownerID 名下二维码会话的最新状态，不重复返回二维码图片。
func (s *QRLogin) Status(ownerID uint, sessionID string) (xyResponse.QRLogin, error) {
	session := s.manager.get(ownerID, sessionID)
	if session == nil {
		return xyResponse.QRLogin{}, errors.New("二维码会话不存在或已过期")
	}
	return session.response(false), nil
}

// Cancel 将 ownerID 名下会话置为取消并停止轮询。
func (s *QRLogin) Cancel(ownerID uint, sessionID string) error {
	session := s.manager.get(ownerID, sessionID)
	if session == nil {
		return errors.New("二维码会话不存在或已过期")
	}
	session.finish(qrStatusCancelled, "登录已取消", "")
	return nil
}

// active 返回指定用户账号尚未终止且未到期的会话。
func (m *QRLoginManager) active(ownerID, accountID uint) *QRSession {
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

// cleanupLocked 清理终态保留期已超过五分钟的会话索引。
func (m *QRLoginManager) cleanupLocked(now time.Time) {
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

// store 同时更新会话主索引和用户账号唯一索引。
func (m *QRLoginManager) store(session *QRSession) {
	m.mu.Lock()
	m.sessions[session.id] = session
	m.byAccount[[2]uint{session.ownerID, session.accountID}] = session.id
	m.mu.Unlock()
}

// get 仅返回归属 ownerID 的二维码会话。
func (m *QRLoginManager) get(ownerID uint, sessionID string) *QRSession {
	m.mu.Lock()
	session := m.sessions[sessionID]
	m.mu.Unlock()
	if session == nil || session.ownerID != ownerID {
		return nil
	}
	return session
}

// poll 周期查询二维码状态，并完成令牌交换、Cookie 刷新和凭据保存。
func (m *QRLoginManager) poll(ctx context.Context, session *QRSession) {
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

// saveCredential 加密保存完整 Cookie，并更新账号的闲鱼用户标识。
func (m *QRLoginManager) saveCredential(ctx context.Context, session *QRSession) error {
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

// response 生成线程安全的会话响应，includeQR 控制是否携带二维码图片。
func (s *QRSession) response(includeQR bool) xyResponse.QRLogin {
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

// terminal 判断会话是否已进入不可变终态。
func (s *QRSession) terminal() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status == qrStatusConfirmed || s.status == qrStatusExpired || s.status == qrStatusCancelled || s.status == qrStatusVerificationRequired || s.status == qrStatusError
}

// setStatus 仅在非终态时更新会话状态和提示。
func (s *QRSession) setStatus(status, message string) {
	s.mu.Lock()
	if !s.isTerminalLocked() {
		s.status = status
		s.message = message
	}
	s.mu.Unlock()
}

// finish 原子写入终态并取消后台轮询。
func (s *QRSession) finish(status, message, verificationURL string) {
	s.mu.Lock()
	if !s.isTerminalLocked() {
		s.status = status
		s.message = message
		s.verificationURL = verificationURL
		s.cancel()
	}
	s.mu.Unlock()
}

// isTerminalLocked 在持有会话锁时判断终态。
func (s *QRSession) isTerminalLocked() bool {
	return s.status == qrStatusConfirmed || s.status == qrStatusExpired || s.status == qrStatusCancelled || s.status == qrStatusVerificationRequired || s.status == qrStatusError
}
