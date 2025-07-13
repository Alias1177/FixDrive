package handler

import (
	"FixDrive/internal/admin"
	"encoding/json"
	"net/http"
)

type Handler struct {
	repo *admin.UserRepo
}

func NewHandler(repo *admin.UserRepo) *Handler {
	return &Handler{repo: repo}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetByEmail(req.Email)
	if err != nil {
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	if user.Password != req.Password {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))
}
