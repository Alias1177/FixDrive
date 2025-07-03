package auth

import (
	"FixDrive/models/client"
	"FixDrive/repo/authRepo"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrUserExists          = errors.New("user already exists")
	ErrInvalidToken        = errors.New("invalid token")
	ErrRefreshTokenInvalid = errors.New("refresh token invalid or expired")
)

// Время жизни токенов
const (
	AccessTokenDuration  = 30 * time.Minute    // 30 минут
	RefreshTokenDuration = 30 * 24 * time.Hour // 30 дней
)

type service struct {
	db        *sqlx.DB
	jwtSecret string
}

func NewService(db *sqlx.DB, jwtSecret string) authRepo.Service {
	return &service{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *service) Register(ctx context.Context, req client.RegisterRequest) (client.TokenResponse, error) {
	// Проверяем существует ли пользователь
	var count int
	err := s.db.GetContext(ctx, &count, "SELECT COUNT(*) FROM client WHERE email = $1", req.Email)
	if err != nil {
		return client.TokenResponse{}, err
	}
	if count > 0 {
		return client.TokenResponse{}, ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return client.TokenResponse{}, err
	}

	// Создаем пользователя
	var userID int64
	query := `INSERT INTO client (email, password_hash, phone_number, first_name, last_name, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, 'active', NOW(), NOW()) RETURNING id`

	err = s.db.QueryRowContext(ctx, query, req.Email, string(hash), req.PhoneNumber, req.FirstName, req.LastName).Scan(&userID)
	if err != nil {
		return client.TokenResponse{}, err
	}

	// Создаем пользователя для ответа
	userInfo := client.UserInfo{
		ID:          userID,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Status:      "active",
	}

	// Генерируем токены
	return s.generateTokenPair(ctx, userInfo)
}

func (s *service) Login(ctx context.Context, req client.LoginRequest) (client.TokenResponse, error) {
	var user client.Client
	err := s.db.GetContext(ctx, &user, "SELECT id, email, password_hash, phone_number, first_name, last_name, status FROM client WHERE email = $1", req.Email)
	if err != nil {
		return client.TokenResponse{}, ErrUserNotFound
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return client.TokenResponse{}, ErrInvalidPassword
	}

	userInfo := client.UserInfo{
		ID:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Status:      user.Status,
	}

	// Генерируем токены
	return s.generateTokenPair(ctx, userInfo)
}

func (s *service) RefreshToken(ctx context.Context, req client.RefreshTokenRequest) (client.TokenResponse, error) {
	// Проверяем refresh token в БД
	var refreshToken client.RefreshToken
	query := `SELECT id, user_id, token, expires_at, is_revoked 
			  FROM refresh_tokens 
			  WHERE token = $1 AND is_revoked = false`

	err := s.db.GetContext(ctx, &refreshToken, query, req.RefreshToken)
	if err != nil {
		return client.TokenResponse{}, ErrRefreshTokenInvalid
	}

	// Проверяем, не истек ли токен
	if time.Now().After(refreshToken.ExpiresAt) {
		return client.TokenResponse{}, ErrRefreshTokenInvalid
	}

	// Отзываем старый refresh token
	_, err = s.db.ExecContext(ctx, "UPDATE refresh_tokens SET is_revoked = true WHERE id = $1", refreshToken.ID)
	if err != nil {
		return client.TokenResponse{}, err
	}

	// Получаем информацию о пользователе
	var user client.Client
	err = s.db.GetContext(ctx, &user, "SELECT id, email, phone_number, first_name, last_name, status FROM client WHERE id = $1", refreshToken.UserID)
	if err != nil {
		return client.TokenResponse{}, ErrUserNotFound
	}

	userInfo := client.UserInfo{
		ID:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Status:      user.Status,
	}

	// Генерируем новые токены
	return s.generateTokenPair(ctx, userInfo)
}

func (s *service) ValidateToken(ctx context.Context, tokenString string) (client.UserInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return client.UserInfo{}, ErrInvalidToken
	}

	if !token.Valid {
		return client.UserInfo{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return client.UserInfo{}, ErrInvalidToken
	}

	// Проверяем тип токена
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return client.UserInfo{}, ErrInvalidToken
	}

	userID := int64(claims["user_id"].(float64))

	var user client.Client
	err = s.db.GetContext(ctx, &user, "SELECT id, email, phone_number, first_name, last_name, status FROM client WHERE id = $1", userID)
	if err != nil {
		return client.UserInfo{}, ErrUserNotFound
	}

	return client.UserInfo{
		ID:          user.ID,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Status:      user.Status,
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	// Отзываем конкретный refresh token
	_, err := s.db.ExecContext(ctx, "UPDATE refresh_tokens SET is_revoked = true WHERE token = $1", refreshToken)
	return err
}

func (s *service) LogoutAll(ctx context.Context, userID int64) error {
	// Отзываем все refresh токены пользователя
	_, err := s.db.ExecContext(ctx, "UPDATE refresh_tokens SET is_revoked = true WHERE user_id = $1", userID)
	return err
}

func (s *service) GetAllUsers(ctx context.Context) ([]client.UserInfo, error) {
	var users []client.UserInfo
	query := `SELECT id, email, phone_number, first_name, last_name, status 
			  FROM client 
			  ORDER BY created_at DESC`

	err := s.db.SelectContext(ctx, &users, query)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *service) generateTokenPair(ctx context.Context, userInfo client.UserInfo) (client.TokenResponse, error) {
	now := time.Now()

	// Создаем Access Token
	accessClaims := jwt.MapClaims{
		"user_id": userInfo.ID,
		"email":   userInfo.Email,
		"type":    "access",
		"exp":     now.Add(AccessTokenDuration).Unix(),
		"iat":     now.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return client.TokenResponse{}, err
	}

	// Генерируем Refresh Token (случайная строка)
	refreshTokenBytes := make([]byte, 32)
	_, err = rand.Read(refreshTokenBytes)
	if err != nil {
		return client.TokenResponse{}, err
	}
	refreshTokenString := hex.EncodeToString(refreshTokenBytes)

	// Сохраняем refresh token в БД
	refreshExpiresAt := now.Add(RefreshTokenDuration)
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at, created_at, is_revoked)
			  VALUES ($1, $2, $3, $4, false)`

	_, err = s.db.ExecContext(ctx, query, userInfo.ID, refreshTokenString, refreshExpiresAt, now)
	if err != nil {
		return client.TokenResponse{}, err
	}

	return client.TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenDuration.Seconds()),
		ExpiresAt:    now.Add(AccessTokenDuration),
		UserInfo:     userInfo,
	}, nil
}
