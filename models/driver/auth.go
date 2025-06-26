package driver

import "time"

// Структуры запросов
type RegisterRequest struct {
	Email             string `json:"email" validate:"required,email"`
	Password          string `json:"password" validate:"required,min=6"`
	PhoneNumber       string `json:"phone_number" validate:"required"`
	FirstName         string `json:"first_name" validate:"required"`
	LastName          string `json:"last_name" validate:"required"`
	LicenseNumber     string `json:"license_number" validate:"required"`
	LicenseExpiryDate string `json:"license_expiry_date" validate:"required"`
	VehicleModel      string `json:"vehicle_model" validate:"required"`
	VehicleNumber     string `json:"vehicle_number" validate:"required"`
	VehicleYear       int    `json:"vehicle_year" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Структуры ответов
type TokenResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int        `json:"expires_in"` // в секундах
	ExpiresAt    time.Time  `json:"expires_at"`
	DriverInfo   DriverInfo `json:"driver_info"`
}

type DriverInfo struct {
	ID                int64   `json:"id" db:"id"`
	Email             string  `json:"email" db:"email"`
	PhoneNumber       string  `json:"phone_number" db:"phone_number"`
	FirstName         string  `json:"first_name" db:"first_name"`
	LastName          string  `json:"last_name" db:"last_name"`
	LicenseNumber     string  `json:"license_number" db:"license_number"`
	LicenseExpiryDate string  `json:"license_expiry_date" db:"license_expiry_date"`
	VehicleModel      string  `json:"vehicle_model" db:"vehicle_model"`
	VehicleNumber     string  `json:"vehicle_number" db:"vehicle_number"`
	VehicleYear       int     `json:"vehicle_year" db:"vehicle_year"`
	Status            string  `json:"status" db:"status"`
	Rating            float64 `json:"rating" db:"rating"`
}

// Модель водителя в БД
type Driver struct {
	ID                int64     `json:"id" db:"id"`
	Email             string    `json:"email" db:"email"`
	PasswordHash      string    `json:"-" db:"password_hash"`
	PhoneNumber       string    `json:"phone_number" db:"phone_number"`
	FirstName         string    `json:"first_name" db:"first_name"`
	LastName          string    `json:"last_name" db:"last_name"`
	LicenseNumber     string    `json:"license_number" db:"license_number"`
	LicenseExpiryDate time.Time `json:"license_expiry_date" db:"license_expiry_date"`
	VehicleModel      string    `json:"vehicle_model" db:"vehicle_model"`
	VehicleNumber     string    `json:"vehicle_number" db:"vehicle_number"`
	VehicleYear       int       `json:"vehicle_year" db:"vehicle_year"`
	Status            string    `json:"status" db:"status"`
	Rating            float64   `json:"rating" db:"rating"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Модель для refresh токенов водителей
type RefreshToken struct {
	ID        int64     `json:"id" db:"id"`
	DriverID  int64     `json:"driver_id" db:"driver_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IsRevoked bool      `json:"is_revoked" db:"is_revoked"`
}
