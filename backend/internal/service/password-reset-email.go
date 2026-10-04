package service

import (
	"fmt"
	"net/smtp"
	"strings"
)

func (s *EmailService) SendPasswordResetEmail(
	to string,
	resetURL string,
) error {

	subject := "Reset your Forensics account password"

	body := fmt.Sprintf(
		`Hello,

We received a request to reset the password for your Forensics account.

Use the following link to reset your password:

%s

This link is valid for a limited time and can only be used once.

If you did not request a password reset, you can safely ignore this email.

Regards,
Forensics Security Team
`,
		resetURL,
	)

	message := strings.Join([]string{
		"From: " + s.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	auth := smtp.PlainAuth(
		"",
		s.username,
		s.password,
		s.host,
	)

	return smtp.SendMail(
		s.host+":"+s.port,
		auth,
		s.from,
		[]string{to},
		[]byte(message),
	)
}
