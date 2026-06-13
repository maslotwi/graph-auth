package api

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/maslotwi/graph-auth/helpers/environment"
)

func SendMagicLinkEmail(to, token string) error {
	link := fmt.Sprintf("%s/verify?token=%s", environment.FrontendURL, token)
	subject := "Your graph-auth login link"
	message := fmt.Sprintf(
		"Subject: %s\r\nFrom: %s\r\nTo: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nHi,\r\n\nClick the link below to sign in. It expires in 15 minutes.\r\n\r\n%s\r\n\r\nIf you did not request this, you can safely ignore this email.\r\n",
		subject, environment.SMTPFrom, to, link,
	)

	addr := environment.SMTPHost + ":" + environment.SMTPPort

	// Port 465 uses implicit TLS; anything else (587, 1025) uses plain SMTP.
	if environment.SMTPPort == "465" {
		return sendWithTLS(addr, to, message)
	}
	return sendPlain(addr, to, message)
}

func sendWithTLS(addr, to, message string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: environment.SMTPServerName})
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	client, err := smtp.NewClient(conn, environment.SMTPHost)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if environment.SMTPUser != "" && environment.SMTPPass != "" {
		auth := smtp.PlainAuth("", environment.SMTPUser, environment.SMTPPass, environment.SMTPHost)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	return sendMessage(client, environment.SMTPFrom, to, message)
}

func sendPlain(addr, to, message string) error {
	var auth smtp.Auth
	if environment.SMTPUser != "" && environment.SMTPPass != "" {
		auth = smtp.PlainAuth("", environment.SMTPUser, environment.SMTPPass, environment.SMTPHost)
	}
	return smtp.SendMail(addr, auth, environment.SMTPFrom, []string{to}, []byte(message))
}

func sendMessage(client *smtp.Client, from, to, message string) error {
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp rcpt: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = fmt.Fprint(w, message); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	return w.Close()
}
