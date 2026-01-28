package config

import (
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`

	SMTP struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		From     string `yaml:"from"`
	} `yaml:"smtp"`

	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	GeoAPI struct {
		Provider string `yaml:"provider"`
		APIKey   string `yaml:"api_key"`
		URL      string `yaml:"url"`
	} `yaml:"geo_api"`

	// New fields for tracking
	App struct {
		Env        string `yaml:"env"`
		BaseURL    string `yaml:"base_url"`
		TrackingID string `yaml:"tracking_id"`
	} `yaml:"app"`
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// Load YAML config first
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	// Override with environment variables
	config.overrideWithEnvVars()

	return config, nil
}

func (c *Config) overrideWithEnvVars() {
	// Server settings
	if port := GetEnv("PORT", ""); port != "" {
		c.Server.Port = port
	}
	if host := GetEnv("HOST", ""); host != "" {
		c.Server.Host = host
	}

	// App settings - these are the key ones for your tracking pixel
	if env := GetEnv("APP_ENV", ""); env != "" {
		c.App.Env = env
	} else if c.App.Env == "" {
		c.App.Env = "development" // Default
	}

	if baseURL := GetEnv("BASE_URL", ""); baseURL != "" {
		c.App.BaseURL = baseURL
	} else if c.App.BaseURL == "" && c.App.Env == "production" {
		// In production without BASE_URL set, you might want to log a warning
		log.Printf("WARNING: BASE_URL not set in production environment")
	}

	if trackingID := GetEnv("TRACKING_ID", ""); trackingID != "" {
		c.App.TrackingID = trackingID
	}

	// Existing overrides (keep your existing structure)
	if smtpHost := GetEnv("SMTP_HOST", ""); smtpHost != "" {
		c.SMTP.Host = smtpHost
	}
	// ... add other existing env overrides as needed
}

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func MustLoadConfig(configPath string) *Config {
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return config
}

func (c *Config) GetBaseURL(requestHost string) string {
	if c.App.BaseURL != "" {
		return c.App.BaseURL
	}

	if c.App.Env == "production" && c.Server.Host != "" {
		// In production, assume HTTPS
		if !strings.HasPrefix(c.Server.Host, "http") {
			return "https://" + c.Server.Host
		}
		return c.Server.Host
	}

	// Default to development URL
	return "http://localhost:" + c.Server.Port
}
