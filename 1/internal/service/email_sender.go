package service

import (
	"context"
	"fmt"
	"net/smtp"
)

type EmailSender struct {
	smtpAddr string
	from     string
}

func NewEmailSender(smtpAddr, from string) *EmailSender {
	return &EmailSender{smtpAddr: smtpAddr, from: from}
}

func (e *EmailSender) Name() string {
	return "email"
}

func (e *EmailSender) Send(ctx context.Context, recipient, message string) error {
	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: Notification\r\n"+
			"\r\n"+
			"%s\r\n",
		e.from, recipient, message,
	))

	err := smtp.SendMail(e.smtpAddr, nil, e.from, []string{recipient}, msg)
	if err != nil {
		return fmt.Errorf("email send failed: %w", err)
	}

	return nil
}
