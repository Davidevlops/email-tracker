package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server struct {
		Port string
		Host string
	}
	SMTP struct {
		Host     string
		Port     int
		Username string
		Password string
		From     string
	}
	Redis struct {
		Host     string
		Port     int
		Password string
		DB       int
	}
	GeoAPI struct {
		Provider string
		APIKey   string
		URL      string
	}
	App struct {
		Env        string
		BaseURL    string
		TrackingID string
	}
}

// LoadConfig reads config directly from environment variables
func LoadConfig() *Config {
	// Load environment variables (optional in production)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg := &Config{}

	// Server
	cfg.Server.Port = getEnv("PORT", "8080")
	cfg.Server.Host = getEnv("HOST", "0.0.0.0")

	// App
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.BaseURL = getEnv("BASE_URL", "")
	cfg.App.TrackingID = getEnv("TRACKING_ID", "dev_track_001")

	// SMTP
	cfg.SMTP.Host = getEnv("SMTP_HOST", "smtp.gmail.com")
	cfg.SMTP.Port = getEnvAsInt("SMTP_PORT", 587)
	cfg.SMTP.Username = getEnv("SMTP_USERNAME", "")
	cfg.SMTP.Password = getEnv("SMTP_PASSWORD", "")
	cfg.SMTP.From = cfg.SMTP.Username

	// Redis
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port = getEnvAsInt("REDIS_PORT", 6379)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = getEnvAsInt("REDIS_DB", 0)

	// Geo API
	cfg.GeoAPI.Provider = getEnv("GEO_PROVIDER", "ip-api")
	cfg.GeoAPI.APIKey = getEnv("GEO_API_KEY", "")
	cfg.GeoAPI.URL = getEnv("GEO_URL", "http://ip-api.com/json/")

	return cfg
}

// Helper: string env
func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

// Helper: int env
func getEnvAsInt(key string, defaultVal int) int {
	if valStr, exists := os.LookupEnv(key); exists {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return defaultVal
}

// GetBaseURL returns correct base URL for email tracking
func (c *Config) GetBaseURL(requestHost string) string {
	if c.App.BaseURL != "" {
		return c.App.BaseURL
	}

	if c.App.Env == "production" && c.Server.Host != "" {
		if !strings.HasPrefix(c.Server.Host, "http") {
			return "https://" + c.Server.Host
		}
		return c.Server.Host
	}

	return "http://localhost:" + c.Server.Port
}

// MustLoadConfig helper
func MustLoadConfig() *Config {
	cfg := LoadConfig()
	if cfg.SMTP.Username == "" || cfg.SMTP.Password == "" {
		log.Println("WARNING: SMTP credentials are missing")
	}
	return cfg
}
