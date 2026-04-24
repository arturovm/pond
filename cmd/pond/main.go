package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/arturovm/pond/internal/api"
	"github.com/arturovm/pond/internal/conf"
	"github.com/arturovm/pond/internal/credentials"
	"github.com/arturovm/pond/internal/database"
	"github.com/arturovm/pond/internal/fetcher"
	"github.com/arturovm/pond/internal/pond"
	"github.com/arturovm/pond/internal/sessions"
	"github.com/arturovm/pond/internal/sources"
	"github.com/arturovm/pond/internal/subscriptions"
	"github.com/arturovm/pond/internal/users"
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
	if err := database.Migrate(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// wire adapters and domain
	accountService := pond.NewAccountService(
		users.NewSQLite(db),
		credentials.NewSQLite(db),
		sessions.NewSQLite(db),
	)
	subscriptionService := pond.NewSubscriptionService(
		fetcher.NewHTTPFetcher(),
		sources.NewSQLite(db),
		subscriptions.NewSQLite(db),
	)

	// start server
	addr := net.JoinHostPort(conf.Addr, fmt.Sprintf("%d", conf.Port))
	slog.Info("server starting", "addr", addr)
	if err := http.ListenAndServe(addr, api.NewRouter(subscriptionService, accountService, accountService, slog.Default())); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
