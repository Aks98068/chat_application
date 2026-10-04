package service

import (
	"fmt"
	"net/smtp"
)

type EmailService struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewEmailService(
	host string,
	port string,
	username string,
	password string,
	from string,
) *EmailService {
	return &EmailService{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *EmailService) SendVerificationEmail(
	to string,
	verificationURL string,
) error {

	subject := "Verify your email address"

	body := fmt.Sprintf(
		"Please verify your email address by visiting this link:\n\n%s\n\n"+
			"This verification link expires in 24 hours.",
		verificationURL,
	)

	message := []byte(
		"From: " + s.from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body,
	)

	auth := smtp.PlainAuth(
		"",
		s.username,
		s.password,
		s.host,
	)

	address := s.host + ":" + s.port

	if err := smtp.SendMail(
		address,
		auth,
		s.from,
		[]string{to},
		message,
	); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
