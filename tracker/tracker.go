package tracker

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"email-tracker/models"
	"email-tracker/utils"
)

type NotificationSender interface {
	SendNotification(ctx context.Context, to []string, subject string, data map[string]interface{}) error
}

type Tracker struct {
	notificationSender NotificationSender
	trackingData       map[string]*models.Email
	trackingEvents     map[string][]*models.TrackingEvent
	pixelTemplate      *template.Template
}

func NewTracker(notificationSender NotificationSender) *Tracker {
	tmpl, err := template.ParseFiles("templates/tracking_pixel.html")
	if err != nil {
		fmt.Printf("Warning: Could not load tracking pixel template: %v\n", err)
	}

	return &Tracker{
		notificationSender: notificationSender,
		trackingData:       make(map[string]*models.Email),
		trackingEvents:     make(map[string][]*models.TrackingEvent),
		pixelTemplate:      tmpl,
	}
}

func (t *Tracker) GenerateTrackingID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (t *Tracker) EmbedTrackingPixel(emailContent, trackingID, baseURL string) (string, error) {
	if t.pixelTemplate == nil {
		return "", fmt.Errorf("tracking pixel template not loaded")
	}

	data := struct {
		BaseURL    string
		TrackingID string
		Timestamp  int64
	}{
		BaseURL:    baseURL,
		TrackingID: trackingID,
		Timestamp:  time.Now().Unix(),
	}

	var pixelHTML bytes.Buffer
	if err := t.pixelTemplate.Execute(&pixelHTML, data); err != nil {
		return "", fmt.Errorf("failed to execute tracking template: %w", err)
	}

	return emailContent + pixelHTML.String(), nil
}

func (t *Tracker) TrackEmailOpen(w http.ResponseWriter, r *http.Request, trackingID, baseURL string) {
	ip := utils.GetClientIP(r)
	userAgent := r.UserAgent()

	geoInfo, err := utils.GetGeoLocation(ip)
	if err != nil {
		fmt.Printf("Error getting geo location: %v\n", err)
	}

	deviceInfo := utils.ParseUserAgent(userAgent)

	var emailID string
	email, exists := t.trackingData[trackingID]
	if exists {
		emailID = email.ID
	}

	event := &models.TrackingEvent{
		ID:         utils.GenerateUUID(),
		TrackingID: trackingID,
		EmailID:    emailID,
		BaseURL:    baseURL,
		IPAddress:  ip,
		UserAgent:  userAgent,
		Country:    geoInfo.Country,
		City:       geoInfo.City,
		Region:     geoInfo.Region,
		ISP:        geoInfo.ISP,
		OpenedAt:   time.Now(),
		DeviceType: deviceInfo.DeviceType,
		Browser:    deviceInfo.Browser,
		OS:         deviceInfo.OS,
	}

	t.trackingEvents[trackingID] = append(t.trackingEvents[trackingID], event)

	fmt.Printf("📧 Email opened - Tracking ID: %s, BaseURL: %s, IP: %s, Location: %s, %s\n",
		trackingID, baseURL, ip, event.City, event.Country)

	// Send notification if needed
	if exists && email.NotifyOnOpen {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := t.notificationSender.SendNotification(ctx, []string{email.NotifyEmail},
			fmt.Sprintf("📧 Email Opened: %s", email.Subject),
			map[string]interface{}{
				"EmailSubject": email.Subject,
				"Recipient":    email.To,
				"OpenedAt":     event.OpenedAt.Format("2006-01-02 15:04:05"),
				"IPAddress":    event.IPAddress,
				"Location":     fmt.Sprintf("%s, %s, %s", event.City, event.Region, event.Country),
				"Device":       event.DeviceType,
				"Browser":      event.Browser,
				"OS":           event.OS,
				"ISP":          event.ISP,
				"TrackingURL":  fmt.Sprintf("%s/track/%s", event.BaseURL, event.TrackingID),
				"BaseURL":      event.BaseURL,
				"Year":         event.OpenedAt.Year(),
			}); err != nil {
			fmt.Printf("Failed to send notification: %v\n", err)
		}
	}

	// Serve tracking pixel
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Write(gifData)
}

var gifData = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61,
	0x01, 0x00, 0x01, 0x00, 0x80, 0x00,
	0x00, 0xff, 0xff, 0xff, 0x00, 0x00,
	0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02,
	0x02, 0x44, 0x01, 0x00, 0x3b,
}

func (t *Tracker) sendNotification(email *models.Email, event *models.TrackingEvent) {
	// Subject for the notification email
	subject := fmt.Sprintf("📧 Email Opened: %s", email.Subject)

	// Prepare template data
	data := map[string]interface{}{
		"EmailSubject": email.Subject,
		"Recipient":    email.To,
		"OpenedAt":     event.OpenedAt.Format("2006-01-02 15:04:05"),
		"IPAddress":    event.IPAddress,
		"Location":     fmt.Sprintf("%s, %s, %s", event.City, event.Region, event.Country),
		"Device":       event.DeviceType,
		"Browser":      event.Browser,
		"OS":           event.OS,
		"ISP":          event.ISP,
		"TrackingURL":  fmt.Sprintf("%s/track/%s", event.BaseURL, event.TrackingID),
		"BaseURL":      event.BaseURL,
		"Year":         event.OpenedAt.Year(),
	}

	// Recipients
	recipients := []string{email.NotifyEmail}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send the notification email
	if err := t.notificationSender.SendNotification(ctx, recipients, subject, data); err != nil {
		fmt.Printf("Failed to send notification: %v\n", err)
	}
}

func (t *Tracker) RegisterEmail(email *models.Email, trackingID string) {
	t.trackingData[trackingID] = email
}

func (t *Tracker) GetTrackingStats(trackingID string) *models.TrackingEvent {
	if events, exists := t.trackingEvents[trackingID]; exists && len(events) > 0 {
		return events[len(events)-1]
	}
	return nil
}

func (t *Tracker) GetAllTrackingEvents(trackingID string) []*models.TrackingEvent {
	if events, exists := t.trackingEvents[trackingID]; exists {
		return events
	}
	return nil
}

func (t *Tracker) CleanupOldEntries(maxAge time.Duration) {
	cutoff := time.Now().Add(-maxAge)

	for id, email := range t.trackingData {
		if email.SentAt.Before(cutoff) {
			delete(t.trackingData, id)
			delete(t.trackingEvents, id)
		}
	}

	for trackingID, events := range t.trackingEvents {
		var recentEvents []*models.TrackingEvent
		for _, event := range events {
			if event.OpenedAt.After(cutoff) {
				recentEvents = append(recentEvents, event)
			}
		}
		t.trackingEvents[trackingID] = recentEvents
	}
}
