package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/arturovm/pond/internal/conf"
	"github.com/arturovm/pond/internal/database"
)

func main() {
	if conf.Help {
		conf.PrintHelp()
		os.Exit(0)
	}

	if conf.Debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
		slog.Debug("debug mode enabled")
	}

	// setup data directory
	dataDir := conf.DataDir()
	slog.Debug("initializing data directory", "path", dataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		slog.Error("failed to initialize data directory", "error", err)
		os.Exit(1)
	}

	// setup database
	db, err := database.Open(dataDir + "/pond.db")
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// run migrations
	migrationsDir := conf.MigrationsDir()
	slog.Debug("running migrations", "path", migrationsDir)
	if err := database.Migrate(db, migrationsDir); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// start server
	mux := http.NewServeMux()

	addr := net.JoinHostPort(conf.Addr, fmt.Sprintf("%d", conf.Port))
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
