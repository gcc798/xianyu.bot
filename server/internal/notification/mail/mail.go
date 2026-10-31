package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/jordan-wright/email"
)

func Send(recipients, subject, body string) error {
	return send(strings.Split(recipients, ","), subject, body)
}

func SendConfigured(subject, body string) error {
	recipients := strings.Split(strings.TrimSuffix(global.GVA_CONFIG.Email.To, ","), ",")
	return send(recipients, subject, body)
}

func send(recipients []string, subject, body string) error {
	config := global.GVA_CONFIG.Email
	var auth smtp.Auth
	if config.IsLoginAuth {
		auth = loginAuth{username: config.From, password: config.Secret}
	} else {
		auth = smtp.PlainAuth("", config.From, config.Secret, config.Host)
	}

	message := email.NewEmail()
	message.From = config.From
	if config.Nickname != "" {
		message.From = fmt.Sprintf("%s <%s>", config.Nickname, config.From)
	}
	message.To = recipients
	message.Subject = subject
	message.HTML = []byte(body)

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)
	if config.IsSSL {
		return message.SendWithTLS(address, auth, &tls.Config{ServerName: config.Host})
	}
	return message.Send(address, auth)
}

type loginAuth struct {
	username string
	password string
}

func (a loginAuth) Start(*smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

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
