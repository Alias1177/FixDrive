package auth

import (
	"FixDrive/models/driver"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

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

func TestDriverService_Register_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	req := driver.RegisterRequest{
		Email:             "driver@example.com",
		Password:          "password123",
		PhoneNumber:       "+1234567890",
		FirstName:         "Mike",
		LastName:          "Driver",
		LicenseNumber:     "DL123456789",
		LicenseExpiryDate: "2025-12-31",
		VehicleModel:      "Toyota Camry",
		VehicleNumber:     "ABC123",
		VehicleYear:       2020,
	}

	// Mock проверки существования email
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM drivers WHERE email = \$1`).
		WithArgs(req.Email).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock проверки лицензии
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM drivers WHERE license_number = \$1`).
		WithArgs(req.LicenseNumber).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock проверки номера машины
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM drivers WHERE vehicle_number = \$1`).
		WithArgs(req.VehicleNumber).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock создания водителя - полный SQL запрос
	mock.ExpectQuery(`INSERT INTO drivers \(email, password_hash, phone_number, first_name, last_name, license_number, license_expiry_date, vehicle_model, vehicle_number, vehicle_year, status, rating, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10, 'pending', 0\.0, NOW\(\), NOW\(\)\) RETURNING id`).
		WithArgs(req.Email, sqlmock.AnyArg(), req.PhoneNumber, req.FirstName, req.LastName,
			req.LicenseNumber, sqlmock.AnyArg(), req.VehicleModel, req.VehicleNumber, req.VehicleYear).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	// Mock создания refresh token
	mock.ExpectExec(`INSERT INTO driver_refresh_tokens \(driver_id, token, expires_at, created_at, is_revoked\) VALUES \(\$1, \$2, \$3, \$4, false\)`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := service.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, req.Email, resp.DriverInfo.Email)
	assert.Equal(t, req.FirstName, resp.DriverInfo.FirstName)
	assert.Equal(t, req.LicenseNumber, resp.DriverInfo.LicenseNumber)
	assert.Equal(t, "pending", resp.DriverInfo.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverService_Register_EmailExists(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	req := driver.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}

	// Mock проверки существования email - уже существует
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM drivers WHERE email = \$1`).
		WithArgs(req.Email).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := service.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrDriverExists, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverService_Register_LicenseExists(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	req := driver.RegisterRequest{
		Email:         "driver@example.com",
		Password:      "password123",
		LicenseNumber: "DL123456789",
	}

	// Mock проверки email - OK
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM drivers WHERE email = \$1`).
		WithArgs(req.Email).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	// Mock проверки лицензии - уже существует
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM drivers WHERE license_number = \$1`).
		WithArgs(req.LicenseNumber).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	_, err := service.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrLicenseExists, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverService_Login_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	req := driver.LoginRequest{
		Email:    "driver@example.com",
		Password: password,
	}

	licenseExpiry := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	// Mock получения водителя
	mock.ExpectQuery(`SELECT id, email, password_hash, phone_number, first_name, last_name, license_number, license_expiry_date, vehicle_model, vehicle_number, vehicle_year, status, rating FROM drivers WHERE email = \$1`).
		WithArgs(req.Email).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "password_hash", "phone_number", "first_name", "last_name",
			"license_number", "license_expiry_date", "vehicle_model", "vehicle_number",
			"vehicle_year", "status", "rating",
		}).AddRow(1, "driver@example.com", string(hash), "+1234567890", "Mike", "Driver",
			"DL123456789", licenseExpiry, "Toyota Camry", "ABC123", 2020, "active", 4.5))

	// Mock создания refresh token
	mock.ExpectExec(`INSERT INTO driver_refresh_tokens \(driver_id, token, expires_at, created_at, is_revoked\) VALUES \(\$1, \$2, \$3, \$4, false\)`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := service.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, req.Email, resp.DriverInfo.Email)
	assert.Equal(t, "DL123456789", resp.DriverInfo.LicenseNumber)
	assert.Equal(t, 4.5, resp.DriverInfo.Rating)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverService_Login_DriverNotFound(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	req := driver.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	// Mock получения водителя - не найден
	mock.ExpectQuery(`SELECT id, email, password_hash, phone_number, first_name, last_name, license_number, license_expiry_date, vehicle_model, vehicle_number, vehicle_year, status, rating FROM drivers WHERE email = \$1`).
		WithArgs(req.Email).
		WillReturnError(sql.ErrNoRows)

	_, err := service.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, ErrDriverNotFound, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverService_RefreshToken_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	refreshToken := "valid_refresh_token"
	req := driver.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}

	// Mock получения refresh token
	mock.ExpectQuery(`SELECT id, driver_id, token, expires_at, is_revoked FROM driver_refresh_tokens WHERE token = \$1 AND is_revoked = false`).
		WithArgs(refreshToken).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "driver_id", "token", "expires_at", "is_revoked",
		}).AddRow(1, 1, refreshToken, time.Now().Add(time.Hour), false))

	// Mock отзыва старого токена
	mock.ExpectExec(`UPDATE driver_refresh_tokens SET is_revoked = true WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	licenseExpiry := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	// Mock получения водителя
	mock.ExpectQuery(`SELECT id, email, phone_number, first_name, last_name, license_number, license_expiry_date, vehicle_model, vehicle_number, vehicle_year, status, rating FROM drivers WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "phone_number", "first_name", "last_name", "license_number",
			"license_expiry_date", "vehicle_model", "vehicle_number", "vehicle_year", "status", "rating",
		}).AddRow(1, "driver@example.com", "+1234567890", "Mike", "Driver", "DL123456789",
			licenseExpiry, "Toyota Camry", "ABC123", 2020, "active", 4.5))

	// Mock создания нового refresh token
	mock.ExpectExec(`INSERT INTO driver_refresh_tokens \(driver_id, token, expires_at, created_at, is_revoked\) VALUES \(\$1, \$2, \$3, \$4, false\)`).
		WithArgs(int64(1), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(2, 1))

	resp, err := service.RefreshToken(context.Background(), req)

	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, refreshToken, resp.RefreshToken)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDriverService_ValidateToken_Success(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	service := NewService(db, "test-secret")

	// Создаем валидный токен
	tokenString, err := createTestToken(1, "driver@example.com", "driver", "test-secret")
	require.NoError(t, err)

	licenseExpiry := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	// Mock получения водителя
	mock.ExpectQuery(`SELECT id, email, phone_number, first_name, last_name, license_number, license_expiry_date, vehicle_model, vehicle_number, vehicle_year, status, rating FROM drivers WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "phone_number", "first_name", "last_name", "license_number",
			"license_expiry_date", "vehicle_model", "vehicle_number", "vehicle_year", "status", "rating",
		}).AddRow(1, "driver@example.com", "+1234567890", "Mike", "Driver", "DL123456789",
			licenseExpiry, "Toyota Camry", "ABC123", 2020, "active", 4.5))

	driverInfo, err := service.ValidateToken(context.Background(), tokenString)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), driverInfo.ID)
	assert.Equal(t, "driver@example.com", driverInfo.Email)
	assert.Equal(t, "DL123456789", driverInfo.LicenseNumber)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// Вспомогательная функция для создания тестового токена
func createTestToken(driverID int64, email, role, secret string) (string, error) {
	claims := map[string]interface{}{
		"driver_id": float64(driverID),
		"email":     email,
		"role":      role,
		"type":      "access",
		"exp":       time.Now().Add(time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	return token.SignedString([]byte(secret))
}
