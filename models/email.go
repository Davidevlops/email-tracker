package models

import "time"

type Email struct {
	ID           string    `json:"id" bson:"id"`
	From         string    `json:"from" bson:"from"`
	To           string    `json:"to" bson:"to"`
	Subject      string    `json:"subject" bson:"subject"`
	Body         string    `json:"body" bson:"body"`
	TrackingID   string    `json:"tracking_id" bson:"tracking_id"`
	SentAt       time.Time `json:"sent_at" bson:"sent_at"`
	NotifyOnOpen bool      `json:"notify_on_open" bson:"notify_on_open"`
	NotifyEmail  string    `json:"notify_email" bson:"notify_email"`
}

type EmailRequest struct {
	To           []string `json:"to" binding:"required"`
	Subject      string   `json:"subject" binding:"required"`
	Body         string   `json:"body" binding:"required"`
	NotifyOnOpen bool     `json:"notify_on_open"`
	NotifyEmail  string   `json:"notify_email"`
}
