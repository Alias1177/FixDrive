package auth

import (
	"FixDrive/models/client"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func setupTestDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "postgres")
	return sqlxDB, mock
}

func TestService_Register_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	req := client.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		PhoneNumber: "+1234567890",
		FirstName:   "John",
		LastName:    "Doe",
	}

	// Mock проверки существования email
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM client WHERE email = \$1`).
		WithArgs(req.Email).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock создания пользователя - исправляем SQL запрос
	mock.ExpectQuery(`INSERT INTO client \(email, password_hash, phone_number, first_name, last_name, status, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, 'active', NOW\(\), NOW\(\)\) RETURNING id`).
		WithArgs(req.Email, sqlmock.AnyArg(), req.PhoneNumber, req.FirstName, req.LastName).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Mock создания refresh token
	mock.ExpectExec(`INSERT INTO refresh_tokens \(user_id, token, expires_at, created_at, is_revoked\) VALUES \(\$1, \$2, \$3, \$4, false\)`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := service.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, req.Email, resp.UserInfo.Email)
	assert.Equal(t, req.FirstName, resp.UserInfo.FirstName)
	assert.Equal(t, req.LastName, resp.UserInfo.LastName)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Register_UserExists(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	req := client.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock проверки существования email - пользователь уже существует
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM client WHERE email = \$1`).
		WithArgs(req.Email).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := service.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrUserExists, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Login_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	// Подготавливаем хеш пароля
	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	req := client.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	}

	// Mock получения пользователя
	mock.ExpectQuery(`SELECT id, email, password_hash, phone_number, first_name, last_name, status FROM client WHERE email = \$1`).
		WithArgs(req.Email).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "password_hash", "phone_number", "first_name", "last_name", "status",
		}).AddRow(1, "test@example.com", string(hash), "+1234567890", "John", "Doe", "active"))

	// Mock создания refresh token
	mock.ExpectExec(`INSERT INTO refresh_tokens \(user_id, token, expires_at, created_at, is_revoked\) VALUES \(\$1, \$2, \$3, \$4, false\)`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := service.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, req.Email, resp.UserInfo.Email)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Login_UserNotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	req := client.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	// Mock получения пользователя - не найден
	mock.ExpectQuery(`SELECT id, email, password_hash, phone_number, first_name, last_name, status FROM client WHERE email = \$1`).
		WithArgs(req.Email).
		WillReturnError(sql.ErrNoRows)

	_, err := service.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Login_InvalidPassword(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	// Подготавливаем хеш правильного пароля
	correctPassword := "correct_password"
	hash, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	req := client.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong_password",
	}

	// Mock получения пользователя
	mock.ExpectQuery(`SELECT id, email, password_hash, phone_number, first_name, last_name, status FROM client WHERE email = \$1`).
		WithArgs(req.Email).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "password_hash", "phone_number", "first_name", "last_name", "status",
		}).AddRow(1, "test@example.com", string(hash), "+1234567890", "John", "Doe", "active"))

	_, err := service.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RefreshToken_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	refreshToken := "valid_refresh_token"
	req := client.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	// Mock получения refresh token
	mock.ExpectQuery(`SELECT id, user_id, token, expires_at, is_revoked FROM refresh_tokens WHERE token = \$1 AND is_revoked = false`).
		WithArgs(refreshToken).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "token", "expires_at", "is_revoked",
		}).AddRow(1, 1, refreshToken, time.Now().Add(time.Hour), false))

	// Mock отзыва старого токена
	mock.ExpectExec(`UPDATE refresh_tokens SET is_revoked = true WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock получения пользователя
	mock.ExpectQuery(`SELECT id, email, phone_number, first_name, last_name, status FROM client WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "phone_number", "first_name", "last_name", "status",
		}).AddRow(1, "test@example.com", "+1234567890", "John", "Doe", "active"))

	// Mock создания нового refresh token
	mock.ExpectExec(`INSERT INTO refresh_tokens \(user_id, token, expires_at, created_at, is_revoked\) VALUES \(\$1, \$2, \$3, \$4, false\)`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(2, 1))

	resp, err := service.RefreshToken(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, refreshToken, resp.RefreshToken) // Новый токен должен отличаться

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_RefreshToken_Invalid(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	req := client.RefreshTokenRequest{
		RefreshToken: "invalid_token",
	}

	// Mock получения refresh token - не найден
	mock.ExpectQuery(`SELECT id, user_id, token, expires_at, is_revoked FROM refresh_tokens WHERE token = \$1 AND is_revoked = false`).
		WithArgs("invalid_token").
		WillReturnError(sql.ErrNoRows)

	_, err := service.RefreshToken(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrRefreshTokenInvalid, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_Logout_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	refreshToken := "token_to_revoke"

	// Mock отзыва токена
	mock.ExpectExec(`UPDATE refresh_tokens SET is_revoked = true WHERE token = \$1`).
		WithArgs(refreshToken).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := service.Logout(context.Background(), refreshToken)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestService_LogoutAll_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	userID := int64(1)

	// Mock отзыва всех токенов пользователя
	mock.ExpectExec(`UPDATE refresh_tokens SET is_revoked = true WHERE user_id = \$1`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 3)) // Отозвали 3 токена

	err := service.LogoutAll(context.Background(), userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
