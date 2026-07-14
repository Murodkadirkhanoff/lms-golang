package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"lms.chashma.uz/enrollment-service/internal/data"
	"lms.chashma.uz/pkg/database"
	"lms.chashma.uz/pkg/env"
	"lms.chashma.uz/pkg/httperr"
	"lms.chashma.uz/pkg/svcclient"
)

const version = "1.0.0"

type config struct {
	port             int
	envName          string
	db               database.Config
	jwtSecret        []byte
	internalKey      string
	trustedOrigins   []string
	courseServiceURL string
}

type application struct {
	config       config
	logger       *slog.Logger
	models       data.Models
	courseClient *svcclient.Client
	httperr.Responder
}

func main() {
	cfg := config{
		port:    env.Int("ENROLLMENT_PORT", 4003),
		envName: env.String("LMS_ENV", "development"),
		db: database.Config{
			DSN:          env.String("ENROLLMENT_DB_DSN", "postgres://lms:devpassword@localhost:5432/lms_enrollment?sslmode=disable"),
			MaxOpenConns: env.Int("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: env.Int("DB_MAX_IDLE_CONNS", 25),
			MaxIdleTime:  env.Duration("DB_MAX_IDLE_TIME", 15*time.Minute),
		},
		jwtSecret:        []byte(env.String("JWT_SECRET", "")),
		internalKey:      env.String("INTERNAL_KEY", ""),
		trustedOrigins:   strings.Fields(env.String("CORS_TRUSTED_ORIGINS", "http://localhost:3000")),
		courseServiceURL: env.String("COURSE_SERVICE_URL", "http://localhost:4002"),
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("service", "enrollment")

	if len(cfg.jwtSecret) == 0 {
		logger.Error("JWT_SECRET must be set")
		os.Exit(1)
	}

	db, err := database.Open(cfg.db)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("database connection pool established")

	err = database.MigrateUp(db, env.String("ENROLLMENT_MIGRATIONS_PATH", "migrations"))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Info("database migrations applied")

	app := &application{
		config:       cfg,
		logger:       logger,
		models:       data.NewModels(db),
		courseClient: svcclient.New(cfg.courseServiceURL, cfg.internalKey),
		Responder:    httperr.Responder{Logger: logger},
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("starting server", "addr", srv.Addr, "env", cfg.envName)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
