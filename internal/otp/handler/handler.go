package handler

import (
	"FixDrive/internal/otp"
	"FixDrive/models"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	otpService *otp.Service
}

func NewHandler(otpService *otp.Service) *Handler {
	return &Handler{
		otpService: otpService,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/send", h.sendOTP)
	r.Post("/verify", h.verifyOTP)

	return r
}

func (h *Handler) sendOTP(w http.ResponseWriter, r *http.Request) {
	var req models.SendOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	if req.Phone == "" {
		http.Error(w, "Номер телефона обязателен", http.StatusBadRequest)
		return
	}

	if err := h.otpService.SendOTP(r.Context(), req.Phone); err != nil {
		http.Error(w, "Ошибка отправки OTP", http.StatusInternalServerError)
		return
	}

	response := models.OTPResponse{
		Success: true,
		Message: "OTP код отправлен",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var req models.VerifyOTPRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	if req.Phone == "" || req.Code == "" {
		http.Error(w, "Номер телефона и код обязательны", http.StatusBadRequest)
		return
	}

	valid, err := h.otpService.VerifyOTP(r.Context(), req.Phone, req.Code)
	if err != nil {
		http.Error(w, "Ошибка проверки OTP", http.StatusInternalServerError)
		return
	}

	if !valid {
		response := models.OTPResponse{
			Success: false,
			Message: "Неверный или истекший код",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.OTPResponse{
		Success: true,
		Message: "Код успешно верифицирован",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
