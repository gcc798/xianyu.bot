package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/gcc798/xianyu.bot/internal/config"
	"github.com/jordan-wright/email"
)

// Sender 使用固定 SMTP 配置发送 HTML 邮件。
type Sender struct {
	config config.Email // config 保存服务器地址与认证凭据，不对外暴露。
}

// NewSender 创建使用 settings 的邮件发送器。
func NewSender(settings config.Email) *Sender { return &Sender{config: settings} }

// Send 将逗号分隔的 recipients 作为收件人发送 HTML 正文。
func (s *Sender) Send(recipients, subject, body string) error {
	return s.send(strings.Split(recipients, ","), subject, body)
}

// send 组装邮件并根据配置选择普通 SMTP 或 TLS 连接。
func (s *Sender) send(recipients []string, subject, body string) error {
	settings := s.config
	var auth smtp.Auth
	if settings.IsLoginAuth {
		auth = loginAuth{username: settings.From, password: settings.Secret}
	} else {
		auth = smtp.PlainAuth("", settings.From, settings.Secret, settings.Host)
	}

	message := email.NewEmail()
	message.From = settings.From
	if settings.Nickname != "" {
		message.From = fmt.Sprintf("%s <%s>", settings.Nickname, settings.From)
	}
	message.To = recipients
	message.Subject = subject
	message.HTML = []byte(body)

	address := fmt.Sprintf("%s:%d", settings.Host, settings.Port)
	if settings.IsSSL {
		return message.SendWithTLS(address, auth, &tls.Config{ServerName: settings.Host})
	}
	return message.Send(address, auth)
}

// loginAuth 实现部分 SMTP 服务要求的 LOGIN 认证流程。
type loginAuth struct {
	username string // username 是认证用户名。
	password string // password 是认证密码。
}

// Start 声明使用 LOGIN 认证机制。
func (a loginAuth) Start(*smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

// Next 根据服务端提示依次返回用户名或密码。
func (a loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	prompt := strings.ToLower(string(fromServer))
	switch {
	case strings.Contains(prompt, "username"), strings.Contains(prompt, "user"):
		return []byte(a.username), nil
	case strings.Contains(prompt, "password"), strings.Contains(prompt, "pass"):
		return []byte(a.password), nil
	default:
		return nil, nil
	}
}
