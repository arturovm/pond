package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/arturovm/pond/internal/pond"
)

type SubscribeHandler struct {
	subscriber pond.Subscriber
}

func NewSubscribeHandler(s pond.Subscriber) *SubscribeHandler {
	return &SubscribeHandler{subscriber: s}
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
	if err := h.subscriber.Subscribe(body.URL); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
