package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"

	"github.com/arturovm/pond/internal/pond"
)

type LoginHandler struct {
	authenticator pond.Authenticator
	logger        *slog.Logger
}

func NewLoginHandler(a pond.Authenticator, logger *slog.Logger) *LoginHandler {
	return &LoginHandler{authenticator: a, logger: logger}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Username == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if body.Password == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var ip netip.Addr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ip, _ = netip.ParseAddr(host)
	}
	token, err := h.authenticator.Authenticate(body.Username, body.Password, ip)
	if err != nil {
		if errors.Is(err, pond.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		h.logger.Error("login failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(struct {
		SessionToken string `json:"session_token"`
	}{SessionToken: hex.EncodeToString(token)})
}
