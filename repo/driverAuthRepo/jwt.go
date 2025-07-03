package driverAuthRepo

import (
	"FixDrive/models/driver"
	"context"
)

type Service interface {
	Register(ctx context.Context, req driver.RegisterRequest) (driver.TokenResponse, error)
	Login(ctx context.Context, req driver.LoginRequest) (driver.TokenResponse, error)
	RefreshToken(ctx context.Context, req driver.RefreshTokenRequest) (driver.TokenResponse, error)
	ValidateToken(ctx context.Context, token string) (driver.DriverInfo, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, driverID int64) error
	GetAllDrivers(ctx context.Context) ([]driver.DriverInfo, error)
}
