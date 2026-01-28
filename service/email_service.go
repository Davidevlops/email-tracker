package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"email-tracker/config"
	"email-tracker/models"
	"email-tracker/notification"
	"email-tracker/tracker"
)

type EmailService struct {
	config   *config.Config
	tracker  *tracker.Tracker
	notifier *notification.Sender
}

func NewEmailService(cfg *config.Config, tr *tracker.Tracker, nt *notification.Sender) *EmailService {
	return &EmailService{
		config:   cfg,
		tracker:  tr,
		notifier: nt,
	}
}

func (s *EmailService) SendTrackedEmail(
	ctx context.Context,
	req *models.EmailRequest,
	baseURL string,
) (string, error) {

	// Generate tracking ID
	trackingID, err := s.tracker.GenerateTrackingID()
	if err != nil {
		return "", fmt.Errorf("failed to generate tracking ID: %w", err)
	}

	// Embed tracking pixel in email body
	trackedBody, err := s.tracker.EmbedTrackingPixel(req.Body, trackingID, baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to embed tracking pixel: %w", err)
	}

	emailCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Send email
	if err := s.notifier.SendEmail(
		emailCtx,
		req.To,
		req.Subject,
		trackedBody,
	); err != nil {
		return "", fmt.Errorf("failed to send email: %w", err)
	}

	// Create and store email model
	emailModel := &models.Email{
		ID:           trackingID,
		From:         s.config.SMTP.From,
		To:           strings.Join(req.To, ","),
		Subject:      req.Subject,
		Body:         req.Body,
		TrackingID:   trackingID,
		SentAt:       time.Now(),
		NotifyOnOpen: req.NotifyOnOpen,
		NotifyEmail:  req.NotifyEmail,
	}

	// Register email for tracking
	s.tracker.RegisterEmail(emailModel, trackingID)

	return trackingID, nil
}

func (s *EmailService) GetTrackingInfo(trackingID string) (*models.TrackingEvent, error) {
	// This would fetch from database in production
	// For now, return nil
	return nil, nil
}

func (s *EmailService) GetEmailStats(trackingID string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Add basic stats
	stats["tracking_id"] = trackingID
	stats["generated_at"] = time.Now()

	return stats, nil
}
