package handler

import (
	"FixDrive/internal/driver/auth"
	"FixDrive/models/driver"
	"FixDrive/repo/driverAuthRepo"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service driverAuthRepo.Service
}

func NewHandler(service driverAuthRepo.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.RefreshToken)
	r.Post("/logout", h.Logout)
	r.Post("/logout-all", h.LogoutAll)
	r.Get("/me", h.GetProfile)
	r.Get("/list", h.GetAllDrivers)
	return r
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req driver.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		switch err {
		case auth.ErrDriverExists:
			http.Error(w, "Driver with this email already exists", http.StatusConflict)
		case auth.ErrLicenseExists:
			http.Error(w, "Driver with this license number already exists", http.StatusConflict)
		case auth.ErrVehicleNumberExists:
			http.Error(w, "Driver with this vehicle number already exists", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req driver.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		if err == auth.ErrDriverNotFound || err == auth.ErrInvalidPassword {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req driver.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := h.service.RefreshToken(r.Context(), req)
	if err != nil {
		if err == auth.ErrRefreshTokenInvalid {
			http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req driver.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.service.Logout(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	token := getToken(r)
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	driverInfo, err := h.service.ValidateToken(r.Context(), token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	err = h.service.LogoutAll(r.Context(), driverInfo.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out from all devices"})
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	token := getToken(r)
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	driverInfo, err := h.service.ValidateToken(r.Context(), token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driverInfo)
}

func (h *Handler) GetAllDrivers(w http.ResponseWriter, r *http.Request) {
	drivers, err := h.service.GetAllDrivers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drivers)
}

func getToken(r *http.Request) string {
	auths := r.Header.Get("Authorization")
	if auths == "" {
		return ""
	}
	parts := strings.Split(auths, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
