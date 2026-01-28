package models

import "time"

type TrackingEvent struct {
	ID         string    `json:"id" bson:"id"`
	TrackingID string    `json:"tracking_id" bson:"tracking_id"`
	BaseURL    string    `json:"base_url" bson:"base_url"`
	EmailID    string    `json:"email_id" bson:"email_id"`
	IPAddress  string    `json:"ip_address" bson:"ip_address"`
	UserAgent  string    `json:"user_agent" bson:"user_agent"`
	Country    string    `json:"country" bson:"country"`
	City       string    `json:"city" bson:"city"`
	Region     string    `json:"region" bson:"region"`
	ISP        string    `json:"isp" bson:"isp"`
	OpenedAt   time.Time `json:"opened_at" bson:"opened_at"`
	DeviceType string    `json:"device_type" bson:"device_type"`
	Browser    string    `json:"browser" bson:"browser"`
	OS         string    `json:"os" bson:"os"`
}

type GeoLocation struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	City    string `json:"city"`
	Region  string `json:"region"`
	ISP     string `json:"isp"`
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
}
