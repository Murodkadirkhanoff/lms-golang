// Package config loads runtime configuration from environment variables.
// It is framework plumbing — no bounded-context knowledge lives here.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the fully-resolved application configuration.
type Config struct {
	Env  string
	Port int

	DB DB

	JWTSecret string
	JWTTTL    time.Duration

	CORSTrustedOrigins []string
	FrontendURL        string
	PublicURL          string
	UploadsDir         string

	MailFrom     string
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string

	RateLimit RateLimit
}

// DB holds database connection settings.
type DB struct {
	DSN          string
	MaxOpenConns int
}

// RateLimit configures the per-IP auth rate limiter.
type RateLimit struct {
	Enabled bool
	RPS     float64
	Burst   int
}

// Load reads configuration from the environment, applying the same defaults
// as the Java application.yml so behaviour matches the previous backend.
func Load() (Config, error) {
	cfg := Config{
		Env:                get("LMS_ENV", "development"),
		Port:               getInt("LMS_PORT", 4000),
		JWTSecret:          get("JWT_SECRET", ""),
		JWTTTL:             getDuration("JWT_TTL", 24*time.Hour),
		CORSTrustedOrigins: strings.Fields(get("CORS_TRUSTED_ORIGINS", "http://localhost:3000")),
		FrontendURL:        get("FRONTEND_URL", "http://localhost:3000"),
		PublicURL:          strings.TrimRight(get("PUBLIC_URL", "http://localhost:4000"), "/"),
		UploadsDir:         get("UPLOADS_DIR", "uploads"),
		MailFrom:           get("MAIL_FROM", "no-reply@lms.chashma.uz"),
		SMTPHost:           get("SMTP_HOST", ""),
		SMTPPort:           getInt("SMTP_PORT", 1025),
		SMTPUsername:       get("SMTP_USERNAME", ""),
		SMTPPassword:       get("SMTP_PASSWORD", ""),
		DB: DB{
			DSN:          dbDSN(),
			MaxOpenConns: getInt("DB_MAX_OPEN_CONNS", 25),
		},
		RateLimit: RateLimit{
			Enabled: getBool("RATE_LIMIT_ENABLED", true),
			RPS:     getFloat("RATE_LIMIT_RPS", 2),
			Burst:   getInt("RATE_LIMIT_BURST", 4),
		},
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET must be set (at least 32 bytes for HS256)")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must be at least 32 bytes for HS256")
	}
	return cfg, nil
}

// dbDSN builds a pgx DSN, honouring either a full LMS_DB_DSN or the discrete
// LMS_DB_* variables used by docker-compose.
func dbDSN() string {
	if dsn := os.Getenv("LMS_DB_DSN"); dsn != "" {
		return dsn
	}
	host := get("LMS_DB_HOST", "localhost")
	port := getInt("LMS_DB_PORT", 5432)
	user := get("LMS_DB_USER", "lms")
	pass := get("LMS_DB_PASSWORD", "devpassword")
	name := get("LMS_DB_NAME", "lms")
	sslmode := get("LMS_DB_SSLMODE", "disable")
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", user, pass, host, port, name, sslmode)
}

func get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			return n
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
