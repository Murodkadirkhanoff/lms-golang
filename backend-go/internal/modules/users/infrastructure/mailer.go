package infrastructure

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/chashma/lms/internal/modules/users/application"
)

// Mailer sends the password-reset email. When SMTP is not configured (empty
// host) the link is logged instead — matching the Java MailService dev flow.
type Mailer struct {
	host        string
	port        int
	username    string
	password    string
	from        string
	frontendURL string
}

// NewMailer builds a Mailer.
func NewMailer(host string, port int, username, password, from, frontendURL string) *Mailer {
	return &Mailer{host: host, port: port, username: username, password: password, from: from, frontendURL: frontendURL}
}

var _ application.Mailer = (*Mailer)(nil)

// SendPasswordReset delivers the reset link. Runs in a goroutine so the HTTP
// response time never leaks whether the address exists.
func (m *Mailer) SendPasswordReset(to, token string) {
	link := m.frontendURL + "/reset-password?token=" + token
	go m.send(to, link)
}

func (m *Mailer) send(to, link string) {
	if strings.TrimSpace(m.host) == "" {
		slog.Info("SMTP not configured — password reset link", "email", to, "link", link)
		return
	}

	body := fmt.Sprintf("Hi,\r\n\r\n"+
		"We received a request to reset the password for your account.\r\n"+
		"Open the link below to choose a new password (valid for 45 minutes):\r\n\r\n"+
		"%s\r\n\r\n"+
		"If you didn't request this, you can safely ignore this email.\r\n", link)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Reset your LearnHub password\r\n\r\n%s",
		m.from, to, body)

	addr := m.host + ":" + strconv.Itoa(m.port)
	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}
	if err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg)); err != nil {
		slog.Error("password reset email failed", "email", to, "err", err)
		return
	}
	slog.Info("password reset email sent", "email", to)
}
