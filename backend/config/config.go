package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config เก็บการตั้งค่าทั้งหมดของแอป
type Config struct {
	ThaidClientID     string
	ThaidClientSecret string
	ThaidWellKnownURL string
	FrontendURL       string
	BackendURL        string
	SessionSecret     string
}

// LoadConfig โหลดการตั้งค่าจาก environment variables
func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		ThaidClientID:     getEnv("THAID_CLIENT_ID", "THAID_CLIENT_ID"),
		ThaidClientSecret: getEnv("THAID_CLIENT_SECRET", ""),
		ThaidWellKnownURL: "https://imauth.bora.dopa.go.th/.well-known/openid-configuration",
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		BackendURL:        getEnv("BACKEND_URL", "https://learn2earn.bde.go.th"),
		SessionSecret:     getEnv("SESSION_SECRET", "your-secret-key-change-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
