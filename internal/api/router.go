package api

import (
	"log/slog"
	"net/http"

	"github.com/arturovm/pond/internal/pond"
)

func NewRouter(s pond.Subscriber, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /subscriptions", NewSubscribeHandler(s, logger))
	return mux
}
