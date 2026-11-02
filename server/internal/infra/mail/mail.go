package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/gcc798/go.xianyu.assistant/server/internal/config"
	"github.com/jordan-wright/email"
)

type Sender struct{ config config.Email }

func NewSender(settings config.Email) *Sender { return &Sender{config: settings} }

func (s *Sender) Send(recipients, subject, body string) error {
	return s.send(strings.Split(recipients, ","), subject, body)
}

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
