package config

import (
	"log"
	"os"
	"strconv"
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

	App struct {
		Env        string `yaml:"env"`
		BaseURL    string `yaml:"base_url"`
		TrackingID string `yaml:"tracking_id"`
	} `yaml:"app"`
}

// LoadConfig reads YAML, expands env vars inside, then overrides with env vars
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// Read YAML file as raw string
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Expand any ${ENV_VAR} placeholders in the YAML
	expanded := os.ExpandEnv(string(data))

	// Decode YAML into struct
	if err := yaml.Unmarshal([]byte(expanded), config); err != nil {
		return nil, err
	}

	// Finally, override important values from environment variables directly
	config.overrideWithEnvVars()

	return config, nil
}

// overrideWithEnvVars ensures production Render env vars take precedence
func (c *Config) overrideWithEnvVars() {
	if port := os.Getenv("PORT"); port != "" {
		c.Server.Port = port
	}
	if host := os.Getenv("HOST"); host != "" {
		c.Server.Host = host
	}

	if env := os.Getenv("APP_ENV"); env != "" {
		c.App.Env = env
	}
	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		c.App.BaseURL = baseURL
	}
	if trackingID := os.Getenv("TRACKING_ID"); trackingID != "" {
		c.App.TrackingID = trackingID
	}

	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		c.SMTP.Host = smtpHost
	}
	if smtpPort := os.Getenv("SMTP_PORT"); smtpPort != "" {
		// Convert string to int safely
		if port, err := strconv.Atoi(smtpPort); err == nil {
			c.SMTP.Port = port
		}
	}
	if smtpUser := os.Getenv("SMTP_USERNAME"); smtpUser != "" {
		c.SMTP.Username = smtpUser
		c.SMTP.From = smtpUser // keep From in sync
	}
	if smtpPass := os.Getenv("SMTP_PASSWORD"); smtpPass != "" {
		c.SMTP.Password = smtpPass
	}
}

// MustLoadConfig is a helper for fatal errors
func MustLoadConfig(configPath string) *Config {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	return cfg
}

// GetBaseURL returns the correct URL depending on env
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
