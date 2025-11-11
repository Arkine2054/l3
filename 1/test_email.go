package main

import (
	"log"
	"net/smtp"
)

func main() {
	smtpAddr := "mailhog:1025"
	from := "no-reply@example.com"
	to := "user@example.com"

	msg := []byte("From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: Test Email\r\n\r\n" +
		"This is a test email from worker.\r\n")

	// Отправляем письмо
	if err := smtp.SendMail(smtpAddr, nil, from, []string{to}, msg); err != nil {
		log.Fatalf("failed to send email: %v", err)
	}

	log.Println("email sent successfully")
}
