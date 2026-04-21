package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/arturovm/pond/internal/pond"
	"github.com/google/uuid"
)

type SubscribeHandler struct {
	subscriber pond.Subscriber
	logger     *slog.Logger
}

func NewSubscribeHandler(s pond.Subscriber, logger *slog.Logger) *SubscribeHandler {
	return &SubscribeHandler{subscriber: s, logger: logger}
}

func (h *SubscribeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.URL == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if u, err := url.ParseRequestURI(body.URL); err != nil || u.Scheme == "" || u.Host == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.subscriber.Subscribe(uuid.UUID{}, body.URL); err != nil {
		h.logger.Error("subscribe failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
