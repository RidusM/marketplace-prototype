package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"gitlab.crja72.ru/golang/2025/spring/course/projects/go7/marketplace/auth-service/internal/config"
	"gopkg.in/gomail.v2"
)

const (
	_defaultRetryTime = 500 * time.Millisecond
)

//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *gomail.Dialer
	sender string
}

func New(config *config.EmailConfig) Mailer {
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)

	return Mailer{
		dialer: dialer,
		sender: config.Sender,
	}
}

func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	const op = "email.Send"

	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	msg := gomail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}
		time.Sleep(_defaultRetryTime)
	}

	return nil
}
