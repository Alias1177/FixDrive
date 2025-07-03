package authRepo

import (
	"FixDrive/models/client"
	"context"
)

type Service interface {
	Register(ctx context.Context, req client.RegisterRequest) (client.TokenResponse, error)
	Login(ctx context.Context, req client.LoginRequest) (client.TokenResponse, error)
	RefreshToken(ctx context.Context, req client.RefreshTokenRequest) (client.TokenResponse, error)
	ValidateToken(ctx context.Context, token string) (client.UserInfo, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID int64) error
	GetAllUsers(ctx context.Context) ([]client.UserInfo, error)
}
