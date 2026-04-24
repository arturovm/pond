package api

import (
	"log/slog"
	"net/http"

	"github.com/arturovm/pond/internal/pond"
)

func NewRouter(s pond.Subscriber, ac pond.AccountCreator, a pond.Authenticator, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /accounts", NewCreateAccountHandler(ac, logger))
	mux.Handle("POST /subscriptions", NewSubscribeHandler(s, logger))
	mux.Handle("POST /sessions", NewLoginHandler(a, logger))
	return mux
}
