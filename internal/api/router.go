package api

import (
	"net/http"

	"github.com/arturovm/pond/internal/pond"
)

func NewRouter(s pond.Subscriber) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /subscriptions", NewSubscribeHandler(s))
	return mux
}
