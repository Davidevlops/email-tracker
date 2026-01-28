package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"email-tracker/models"
)

type DeviceInfo struct {
	DeviceType string
	Browser    string
	OS         string
}

func GetClientIP(r *http.Request) string {
	// 1. Cloudflare / some CDNs / modern proxies sometimes use this
	if cf := r.Header.Get("CF-Connecting-IP"); cf != "" {
		if ip := net.ParseIP(cf); ip != nil {
			return ip.String()
		}
	}

	// 2. X-Real-IP  (set by nginx/apache when configured with real_ip module)
	if real := r.Header.Get("X-Real-IP"); real != "" {
		if ip := net.ParseIP(real); ip != nil {
			return ip.String()
		}
	}

	// 3. X-Forwarded-For – take the RIGHTMOST non-private IP
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			ipStr := strings.TrimSpace(parts[i])
			if ip := net.ParseIP(ipStr); ip != nil {
				// Skip private/reserved ranges (very rough check)
				if !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsMulticast() {
					return ipStr
				}
			}
		}
	}

	// 4. Fallback – direct connection or no proxy headers
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func GetGeoLocation(ip string) (*models.GeoLocation, error) {
	// ip-api.com
	if location, err := getGeoFromIPAPI(ip); err == nil {
		return location, nil
	}

	// Fallback: Return basic info
	return &models.GeoLocation{
		IP: ip,
	}, fmt.Errorf("could not determine location")
}

func getGeoFromIPAPI(ip string) (*models.GeoLocation, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ip)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Status  string  `json:"status"`
		Country string  `json:"country"`
		Region  string  `json:"regionName"`
		City    string  `json:"city"`
		ISP     string  `json:"isp"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	if data.Status != "success" {
		return nil, fmt.Errorf("API returned non-success status")
	}
	fmt.Println("data:")
	return &models.GeoLocation{
		IP:      ip,
		Country: data.Country,
		City:    data.City,
		Region:  data.Region,
		ISP:     data.ISP,
		Lat:     fmt.Sprintf("%f", data.Lat),
		Lon:     fmt.Sprintf("%f", data.Lon),
	}, nil
}

func ParseUserAgent(userAgent string) *DeviceInfo {
	info := &DeviceInfo{
		DeviceType: "Desktop",
		Browser:    "Unknown",
		OS:         "Unknown",
	}

	ua := strings.ToLower(userAgent)

	// Detect device type
	if strings.Contains(ua, "mobile") {
		info.DeviceType = "Mobile"
	} else if strings.Contains(ua, "tablet") {
		info.DeviceType = "Tablet"
	}

	// Detect browser
	switch {
	case strings.Contains(ua, "chrome"):
		info.Browser = "Chrome"
	case strings.Contains(ua, "firefox"):
		info.Browser = "Firefox"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		info.Browser = "Safari"
	case strings.Contains(ua, "edge"):
		info.Browser = "Edge"
	case strings.Contains(ua, "opera"):
		info.Browser = "Opera"
	}

	// Detect OS
	switch {
	case strings.Contains(ua, "windows"):
		info.OS = "Windows"
	case strings.Contains(ua, "mac os"):
		info.OS = "macOS"
	case strings.Contains(ua, "linux"):
		info.OS = "Linux"
	case strings.Contains(ua, "android"):
		info.OS = "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		info.OS = "iOS"
	}

	return info
}
