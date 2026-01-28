package notification

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"time"

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

func (s *Sender) SendNotification(
	ctx context.Context,
	to []string,
	subject string,
	data map[string]interface{},
) error {
	// 1. Load HTML template
	// Optimization: In a production app, you should parse templates
	// ONCE at startup and store them in the s.Sender struct.
	tmpl, err := template.ParseFiles("templates/notification.html")
	if err != nil {
		return fmt.Errorf("could not find or parse template file: %w", err)
	}

	// 2. Execute template into a buffer
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to inject data into template: %w", err)
	}

	// 3. Create email message
	e := email.NewEmail()
	e.From = s.config.SMTP.From
	e.To = to
	e.Subject = subject
	e.HTML = body.Bytes()

	// 4. Setup Authentication (Ensure you use a Gmail App Password)
	auth := smtp.PlainAuth(
		"",
		s.config.SMTP.Username,
		s.config.SMTP.Password,
		s.config.SMTP.Host,
	)

	addr := fmt.Sprintf("%s:%d", s.config.SMTP.Host, s.config.SMTP.Port)

	// 5. Send with concurrency-safe timeout
	sendErr := make(chan error, 1)

	go func() {
		// SendWithStartTLS is best for Gmail Port 587
		sendErr <- e.SendWithStartTLS(
			addr,
			auth,
			&tls.Config{
				ServerName: s.config.SMTP.Host,
				MinVersion: tls.VersionTLS12,
			},
		)
	}()

	// 6. Wait for either success, failure, or context cancellation
	select {
	case <-ctx.Done():
		return fmt.Errorf("notification cancelled by context: %w", ctx.Err())

	case err := <-sendErr:
		if err != nil {
			// This is where "Invalid Credentials" will be caught
			return fmt.Errorf("smtp dispatch failed: %w", err)
		}
	}

	return nil
}

// EmailService interface to avoid circular dependency
type EmailService interface {
	GenerateTrackingID() (string, error)
	EmbedTrackingPixel(emailContent, trackingID string) (string, error)
	RegisterEmail(email *models.Email, trackingID string)
}

func (s *Sender) SendEmail(
	ctx context.Context,
	to []string,
	subject, body string,
) error {
	// Build email
	e := email.NewEmail()
	e.From = s.config.SMTP.From
	e.To = to
	e.Subject = subject
	e.HTML = []byte(body)

	addr := fmt.Sprintf("%s:%d", s.config.SMTP.Host, s.config.SMTP.Port)

	// Note: Gmail requires the host in PlainAuth to match the server address
	auth := smtp.PlainAuth(
		"",
		s.config.SMTP.Username,
		s.config.SMTP.Password,
		s.config.SMTP.Host,
	)

	// Context for the entire operation
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	errCh := make(chan error, 1)

	// Run the send operation in a goroutine so the select block
	// can actually catch a timeout if the network hangs.
	go func() {
		errCh <- e.SendWithStartTLS(
			addr,
			auth,
			&tls.Config{
				ServerName: s.config.SMTP.Host,
				// InsecureSkipVerify: true, // Only use for local testing
			},
		)
	}()

	select {
	case <-timeoutCtx.Done():
		return fmt.Errorf("email send timed out: %w", timeoutCtx.Err())

	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("smtp authentication/sending failed: %w", err)
		}
	}

	return nil
}
