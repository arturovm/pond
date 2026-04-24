package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/arturovm/pond/internal/pond"
)

type CreateAccountHandler struct {
	accountCreator pond.AccountCreator
	logger         *slog.Logger
}

func NewCreateAccountHandler(ac pond.AccountCreator, logger *slog.Logger) *CreateAccountHandler {
	return &CreateAccountHandler{accountCreator: ac, logger: logger}
}

func (h *CreateAccountHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	token, err := h.accountCreator.CreateAccount(body.Username, body.Password, parseClientIP(r))
	if err != nil {
		if errors.Is(err, pond.ErrUsernameTaken) {
			http.Error(w, "username taken", http.StatusConflict)
			return
		}
		h.logger.Error("create account failed", "error", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeTokenResponse(w, token)
}
