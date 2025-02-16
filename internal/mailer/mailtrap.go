package mailer

import (
	"fmt"
	gomail "gopkg.in/mail.v2"
	"log"
	"time"
)

type MailtrapMailer struct {
	fromEmail string
	username  string
	password  string
	client    *gomail.Dialer
}

func NewMailtrapMailer(fromEmail, username, password string) *MailtrapMailer {
	return &MailtrapMailer{
		fromEmail: fromEmail,
		client:    gomail.NewDialer("sandbox.smtp.mailtrap.io", 2525, username, password),
	}
}

func (m *MailtrapMailer) Send(username, email string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Hello from the Mailtrap team")
	message.SetBody("text/plain", "This is the Test Body")

	for i := 0; i < MaxRetry; i++ {
		if err := m.client.DialAndSend(message); err != nil {
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		log.Println("Successfully sent email to", email)
		return nil
	}

	return fmt.Errorf("failed to send email after %d attempts", MaxRetry)
}
