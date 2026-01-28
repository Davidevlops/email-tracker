package notification

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"

	"email-tracker/config"
	"email-tracker/models"

	"github.com/jordan-wright/email"
)

type Sender struct {
	config *config.Config
}

func NewSender(cfg *config.Config) *Sender {
	return &Sender{
		config: cfg,
	}
}

func (s *Sender) SendNotification(ctx context.Context, to []string, subject string, data map[string]interface{}) error {
	// Load HTML template
	tmpl, err := template.ParseFiles("templates/notification.html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template with data
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Create email
	e := email.NewEmail()
	e.From = s.config.SMTP.From
	e.To = to
	e.Subject = subject
	e.HTML = body.Bytes()

	// Send email
	auth := smtp.PlainAuth("", s.config.SMTP.Username, s.config.SMTP.Password, s.config.SMTP.Host)

	addr := fmt.Sprintf("%s:%d", s.config.SMTP.Host, s.config.SMTP.Port)

	if err := e.Send(addr, auth); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// EmailService interface to avoid circular dependency
type EmailService interface {
	GenerateTrackingID() (string, error)
	EmbedTrackingPixel(emailContent, trackingID string) (string, error)
	RegisterEmail(email *models.Email, trackingID string)
}

func (s *Sender) SendEmail(ctx context.Context, to []string, subject, body string) error {

	// Create email
	e := email.NewEmail()
	e.From = s.config.SMTP.From
	e.To = to
	e.Subject = subject
	e.HTML = []byte(body)

	// Send email
	auth := smtp.PlainAuth("", s.config.SMTP.Username, s.config.SMTP.Password, s.config.SMTP.Host)
	addr := fmt.Sprintf("%s:%d", s.config.SMTP.Host, s.config.SMTP.Port)

	if err := e.Send(addr, auth); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
