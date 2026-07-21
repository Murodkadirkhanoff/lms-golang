// Package app is the composition root. It wires the platform kernel and the
// bounded-context modules together, honouring the dependency DAG
// (users → courses → enrollment) so there is no import or construction cycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/chashma/lms/internal/modules/courses"
	"github.com/chashma/lms/internal/modules/enrollment"
	"github.com/chashma/lms/internal/modules/users"
	"github.com/chashma/lms/internal/platform/config"
	"github.com/chashma/lms/internal/platform/database"
	"github.com/chashma/lms/internal/platform/web"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Run loads configuration, connects to Postgres, runs each module's
// migrations, wires the modules, and serves HTTP until interrupted.
func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DB.DSN, cfg.DB.MaxOpenConns)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer pool.Close()

	if err := runMigrations(cfg.DB.DSN); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	handler, err := buildHandler(cfg, pool)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting server", "addr", srv.Addr, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

// buildHandler constructs the modules and the HTTP router.
func buildHandler(cfg config.Config, pool *pgxpool.Pool) (http.Handler, error) {
	tokenMaker := web.NewTokenMaker(cfg.JWTSecret, cfg.JWTTTL)
	rateLimiter := web.NewRateLimiter(cfg.RateLimit.Enabled, cfg.RateLimit.RPS, cfg.RateLimit.Burst)

	// Construction order follows the contract DAG.
	usersMod := users.New(pool, tokenMaker, users.MailConfig{
		Host:        cfg.SMTPHost,
		Port:        cfg.SMTPPort,
		Username:    cfg.SMTPUsername,
		Password:    cfg.SMTPPassword,
		From:        cfg.MailFrom,
		FrontendURL: cfg.FrontendURL,
	})
	coursesMod := courses.New(pool, usersMod.Directory())
	enrollMod := enrollment.New(pool, coursesMod.Catalog(), usersMod.Directory())

	up, err := newUploads(cfg.UploadsDir, cfg.PublicURL)
	if err != nil {
		return nil, fmt.Errorf("uploads: %w", err)
	}

	return newRouter(cfg, tokenMaker, rateLimiter, up, func(r chi.Router) {
		usersMod.RegisterRoutes(r, coursesMod.Catalog(), enrollMod.Gate())
		coursesMod.RegisterRoutes(r, enrollMod.Gate())
		enrollMod.RegisterRoutes(r)
	}), nil
}

// runMigrations applies every module's embedded migrations under its own
// version table (database-per-module readiness).
func runMigrations(dsn string) error {
	modules := []struct {
		name string
		fn   func() (fs.FS, string, string)
	}{
		{"users", users.Migrations},
		{"courses", courses.Migrations},
		{"enrollment", enrollment.Migrations},
	}
	for _, m := range modules {
		fsys, dir, table := m.fn()
		if err := database.Migrate(dsn, fsys, dir, table); err != nil {
			return fmt.Errorf("%s: %w", m.name, err)
		}
		slog.Info("migrations applied", "module", m.name)
	}
	return nil
}
