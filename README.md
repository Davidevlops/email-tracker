
# Email Tracker

Email Tracker is a Go-based service for tracking email opens, gathering geolocation and device analytics, and sending real-time notifications when emails are opened. It is designed with extensibility, security, and clean architecture in mind.

---

## Table of Contents
- [Features](#features)
- [Architecture Overview](#architecture-overview)
- [Setup & Installation](#setup--installation)
- [Usage Guide](#usage-guide)
- [API Endpoints](#api-endpoints)
- [Screenshots](#screenshots)
- [Enhancements & Roadmap](#enhancements--roadmap)

---

## Features

- **Email Tracking**: Embeds invisible tracking pixels in outgoing emails to detect when and where emails are opened.
- **IP Geolocation**: Uses multiple free APIs to enrich open events with location data (country, city, ISP, etc).
- **Device & Browser Detection**: Parses user-agent strings to identify device type, browser, and OS.
- **Notification System**: Sends detailed notifications (via email) when tracked emails are opened.
- **Security**: Extracts real client IPs behind proxies/CDNs, sanitizes input, and masks sensitive data.
- **Scalable & Extensible**: Modular codebase with clear separation of concerns for easy extension.

---

## Architecture Overview

The system is organized into the following main components:

- **main.go**: Entry point, server setup, and route definitions.
- **config/**: Configuration loading from environment variables.
- **models/**: Data models for emails and tracking events.
- **service/**: Business logic for sending tracked emails.
- **tracker/**: Tracking pixel generation, event logging, and notification triggers.
- **notification/**: Email notification sender and template rendering.
- **utils/**: Helper functions for geolocation, device parsing, and validation.
- **templates/**: HTML templates for emails, notifications, and tracking pixels.

---

## Setup & Installation

1. **Clone the repository:**
	```sh
	git clone <repo-url>
	cd email-tracker
	```
2. **Configure environment variables:**
	- Copy `.env.example` to `.env` and fill in SMTP and API keys as needed.
3. **Install dependencies:**
	```sh
	go mod tidy
	```
4. **Run the server:**
	```sh
	go run main.go
	```
5. **Server will start on the configured port (default: 8080).**

---

## Usage Guide

### 1. Sending a Tracked Email
Send a POST request to `/api/send-email` with the following JSON body:

```json
{
  "to": ["recipient@example.com"],
  "subject": "Test Email",
  "body": "Hello, this is a tracked email!",
  "notify_on_open": true,
  "notify_email": "your@email.com"
}
```

The system will embed a tracking pixel and send the email. When the recipient opens the email, an open event is logged and a notification is sent (if enabled).

### 2. Tracking Pixel
The tracking pixel is a 1x1 transparent image embedded in the email body:

```html
<img src="https://your-domain/track/{tracking_id}?t={timestamp}" width="1" height="1" style="display:none;" />
```

### 3. Viewing Tracking Stats
GET `/api/tracking/{tracking_id}` returns JSON with open events, geolocation, device, and browser info.

### 4. Dashboard
Visit `/dashboard` for a web-based overview (basic template, can be extended).

---

## API Endpoints

| Method | Endpoint                | Description                       |
|--------|-------------------------|-----------------------------------|
| GET    | `/health`               | Health check                      |
| POST   | `/api/send-email`       | Send a tracked email              |
| GET    | `/track/{id}`           | Tracking pixel endpoint           |
| GET    | `/api/tracking/{id}`    | Get tracking stats for an email   |
| GET    | `/dashboard`            | Dashboard (HTML)                  |

---

## Screenshots

### 1. Dashboard Overview
![Dashboard](Screenshot%202026-02-26%20230133.png)

### 2. Email Sent Confirmation
![Email Sent](Screenshot%202026-02-26%20230254.png)

### 3. Tracking Pixel Request
![Tracking Pixel](Screenshot%202026-02-26%20230324.png)

### 4. Notification Email Example
![Notification Email](Screenshot%202026-02-26%20230408.png)

### 5. Tracking Event Details
![Tracking Event](Screenshot%202026-02-26%20230440.png)

### 6. Geolocation & Device Analytics
![Geo & Device](Screenshot%202026-02-27%20111209.png)

### 7. API Usage Example
![API Usage](Screenshot%202026-02-27%20115410.png)

---

## Enhancements & Roadmap

- Add database persistence (PostgreSQL/Redis)
- Implement authentication for the API
- Add rate limiting
- Create a richer web dashboard
- Add more detailed analytics and reporting
- Support for multiple SMTP providers
- Webhook support for real-time notifications

