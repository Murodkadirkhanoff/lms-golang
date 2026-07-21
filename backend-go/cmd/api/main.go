// Command api is the LMS modular-monolith entrypoint. It runs all bounded
// contexts in one process today; each module can be extracted into its own
// service without touching module code.
package main

import (
	"log/slog"
	"os"

	"github.com/chashma/lms/internal/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	if err := app.Run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}
