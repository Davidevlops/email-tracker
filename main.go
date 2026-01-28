package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email-tracker/config"
	"email-tracker/models"
	"email-tracker/notification"
	"email-tracker/service"
	"email-tracker/tracker"
	"email-tracker/utils"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router       *gin.Engine
	config       *config.Config
	tracker      *tracker.Tracker
	notifier     *notification.Sender
	emailService *service.EmailService
	server       *http.Server
}

func NewServer(cfg *config.Config) *Server {
	// Set Gin mode based on environment
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.Default()

	// Initialize notification sender
	notifier := notification.NewSender(cfg)

	// Initialize tracker
	emailTracker := tracker.NewTracker(notifier)

	// Initialize email service with config
	emailService := service.NewEmailService(cfg, emailTracker, notifier)

	// Clean up old entries periodically
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			emailTracker.CleanupOldEntries(30 * 24 * time.Hour) // 30 days
		}
	}()

	// Log environment info
	log.Printf("Starting server in %s mode", cfg.App.Env)
	if cfg.App.BaseURL != "" {
		log.Printf("BaseURL configured: %s", cfg.App.BaseURL)
	} else {
		log.Printf("BaseURL will be determined dynamically from requests")
	}

	return &Server{
		router:       router,
		config:       cfg,
		tracker:      emailTracker,
		notifier:     notifier,
		emailService: emailService,
	}
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", s.healthCheck)

	// Track email opens
	s.router.GET("/track/:id", s.trackEmailOpen)

	// Send email with tracking
	s.router.POST("/api/send-email", s.sendEmail)

	// Get tracking statistics
	s.router.GET("/api/tracking/:id", s.getTrackingInfo)

	// Dashboard
	s.router.GET("/dashboard", s.dashboard)

	// Static files
	s.router.Static("/static", "./static")

	// Add middleware for dynamic BaseURL
	s.router.Use(s.baseURLMiddleware())
}

// Middleware to inject BaseURL into context
func (s *Server) baseURLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get BaseURL dynamically based on request
		baseURL := s.config.GetBaseURL(c.Request.Host)

		// Store it in context for use in handlers/templates
		c.Set("baseURL", baseURL)

		c.Next()
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	// Get BaseURL from context
	baseURL, _ := c.Get("baseURL")

	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"service":     "email-tracker",
		"version":     "1.0.0",
		"environment": s.config.App.Env,
		"base_url":    baseURL,
		"tracking_id": s.config.App.TrackingID,
	})
}

func (s *Server) trackEmailOpen(c *gin.Context) {
	trackingID := c.Param("id")

	// Get BaseURL for this request
	baseURL, _ := c.Get("baseURL")

	// Pass BaseURL to tracker if needed
	s.tracker.TrackEmailOpen(c.Writer, c.Request, trackingID, baseURL.(string))
}

func (s *Server) sendEmail(c *gin.Context) {
	var req models.EmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate email addresses
	for _, email := range req.To {
		if !utils.ValidateEmail(email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid email: %s", email)})
			return
		}
	}

	// Get BaseURL from context to use in tracking pixel
	baseURL, _ := c.Get("baseURL")

	// Send email using service with BaseURL
	trackingID, err := s.emailService.SendTrackedEmail(c.Request.Context(), &req, baseURL.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Email sent successfully",
		"tracking_id": trackingID,
		"base_url":    baseURL,
		"environment": s.config.App.Env,
	})
}

func (s *Server) getTrackingInfo(c *gin.Context) {
	trackingID := c.Param("id")
	stats := s.tracker.GetTrackingStats(trackingID)

	if stats == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tracking data not found"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (s *Server) dashboard(c *gin.Context) {
	// Get BaseURL from context
	baseURL, _ := c.Get("baseURL")

	// Serve dashboard HTML with BaseURL injected
	c.HTML(http.StatusOK, "./templates/dashboard.html", gin.H{
		"title":       "Email Tracker Dashboard",
		"baseURL":     baseURL,
		"environment": s.config.App.Env,
		"trackingID":  s.config.App.TrackingID,
	})
}

// Helper function to get dynamic BaseURL for templates
func (s *Server) getDynamicBaseURL(c *gin.Context) string {
	baseURL, exists := c.Get("baseURL")
	if exists {
		return baseURL.(string)
	}
	// Fallback to config method
	return s.config.GetBaseURL(c.Request.Host)
}

func (s *Server) Start() error {

	// Add middleware for dynamic BaseURL FIRST
	s.router.Use(s.baseURLMiddleware())
	s.setupRoutes()

	addr := fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on %s", addr)
	log.Printf("Environment: %s", s.config.App.Env)
	log.Printf("Tracking ID: %s", s.config.App.TrackingID)

	if s.config.App.BaseURL != "" {
		log.Printf("Static BaseURL: %s", s.config.App.BaseURL)
	} else {
		log.Printf("Using dynamic BaseURL from requests")
	}

	// Graceful shutdown
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func main() {
	// Load configuration
	cfg := config.MustLoadConfig()

	// Log config status
	log.Printf("Configuration loaded successfully")
	log.Printf("Environment: %s", cfg.App.Env)

	// Create server
	server := NewServer(cfg)

	// Start server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
