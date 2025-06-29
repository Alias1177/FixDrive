package auth

import (
	"FixDrive/models/driver"
	"FixDrive/repo/driverAuthRepo"
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
	ErrDriverNotFound      = errors.New("driver not found")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrDriverExists        = errors.New("driver already exists")
	ErrInvalidToken        = errors.New("invalid token")
	ErrTokenExpired        = errors.New("token expired")
	ErrRefreshTokenInvalid = errors.New("refresh token invalid or expired")
	ErrLicenseExists       = errors.New("license number already exists")
	ErrVehicleNumberExists = errors.New("vehicle number already exists")
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

func NewService(db *sqlx.DB, jwtSecret string) driverAuthRepo.Service {
	return &service{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

func (s *service) Register(ctx context.Context, req driver.RegisterRequest) (driver.TokenResponse, error) {
	// Проверяем существует ли водитель по email
	var emailCount int
	err := s.db.GetContext(ctx, &emailCount, "SELECT COUNT(*) FROM drivers WHERE email = $1", req.Email)
	if err != nil {
		return driver.TokenResponse{}, err
	}
	if emailCount > 0 {
		return driver.TokenResponse{}, ErrDriverExists
	}

	// Проверяем уникальность номера лицензии
	var licenseCount int
	err = s.db.GetContext(ctx, &licenseCount, "SELECT COUNT(*) FROM drivers WHERE license_number = $1", req.LicenseNumber)
	if err != nil {
		return driver.TokenResponse{}, err
	}
	if licenseCount > 0 {
		return driver.TokenResponse{}, ErrLicenseExists
	}

	// Проверяем уникальность номера машины
	var vehicleCount int
	err = s.db.GetContext(ctx, &vehicleCount, "SELECT COUNT(*) FROM drivers WHERE vehicle_number = $1", req.VehicleNumber)
	if err != nil {
		return driver.TokenResponse{}, err
	}
	if vehicleCount > 0 {
		return driver.TokenResponse{}, ErrVehicleNumberExists
	}

	// Хешируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return driver.TokenResponse{}, err
	}

	// Парсим дату истечения лицензии
	licenseExpiry, err := time.Parse("2006-01-02", req.LicenseExpiryDate)
	if err != nil {
		return driver.TokenResponse{}, fmt.Errorf("invalid license expiry date format: %v", err)
	}

	// Создаем водителя
	var driverID int64
	query := `INSERT INTO drivers (email, password_hash, phone_number, first_name, last_name, 
			  license_number, license_expiry_date, vehicle_brand, vehicle_model, vehicle_number, vehicle_year, 
			  status, rating, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'pending', 0.0, NOW(), NOW()) 
			  RETURNING id`

	err = s.db.QueryRowContext(ctx, query, req.Email, string(hash), req.PhoneNumber,
		req.FirstName, req.LastName, req.LicenseNumber, licenseExpiry, req.VehicleBrand, req.VehicleModel,
		req.VehicleNumber, req.VehicleYear).Scan(&driverID)
	if err != nil {
		return driver.TokenResponse{}, err
	}

	// Создаем информацию о водителе для ответа
	driverInfo := driver.DriverInfo{
		ID:                driverID,
		Email:             req.Email,
		PhoneNumber:       req.PhoneNumber,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		LicenseNumber:     req.LicenseNumber,
		LicenseExpiryDate: req.LicenseExpiryDate,
		VehicleBrand:      req.VehicleBrand,
		VehicleModel:      req.VehicleModel,
		VehicleNumber:     req.VehicleNumber,
		VehicleYear:       req.VehicleYear,
		Status:            "pending",
		Rating:            0.0,
	}

	// Генерируем токены
	return s.generateTokenPair(ctx, driverInfo)
}

func (s *service) Login(ctx context.Context, req driver.LoginRequest) (driver.TokenResponse, error) {
	var driverRecord driver.Driver
	err := s.db.GetContext(ctx, &driverRecord,
		`SELECT id, email, password_hash, phone_number, first_name, last_name, 
		 license_number, license_expiry_date, vehicle_brand, vehicle_model, vehicle_number, vehicle_year, 
		 status, rating FROM drivers WHERE email = $1`, req.Email)
	if err != nil {
		return driver.TokenResponse{}, ErrDriverNotFound
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(driverRecord.PasswordHash), []byte(req.Password)); err != nil {
		return driver.TokenResponse{}, ErrInvalidPassword
	}

	driverInfo := driver.DriverInfo{
		ID:                driverRecord.ID,
		Email:             driverRecord.Email,
		PhoneNumber:       driverRecord.PhoneNumber,
		FirstName:         driverRecord.FirstName,
		LastName:          driverRecord.LastName,
		LicenseNumber:     driverRecord.LicenseNumber,
		LicenseExpiryDate: driverRecord.LicenseExpiryDate.Format("2006-01-02"),
		VehicleBrand:      driverRecord.VehicleBrand,
		VehicleModel:      driverRecord.VehicleModel,
		VehicleNumber:     driverRecord.VehicleNumber,
		VehicleYear:       driverRecord.VehicleYear,
		Status:            driverRecord.Status,
		Rating:            driverRecord.Rating,
	}

	// Генерируем токены
	return s.generateTokenPair(ctx, driverInfo)
}

func (s *service) RefreshToken(ctx context.Context, req driver.RefreshTokenRequest) (driver.TokenResponse, error) {
	// Проверяем refresh token в БД
	var refreshToken driver.RefreshToken
	query := `SELECT id, driver_id, token, expires_at, is_revoked 
			  FROM driver_refresh_tokens 
			  WHERE token = $1 AND is_revoked = false`

	err := s.db.GetContext(ctx, &refreshToken, query, req.RefreshToken)
	if err != nil {
		return driver.TokenResponse{}, ErrRefreshTokenInvalid
	}

	// Проверяем, не истек ли токен
	if time.Now().After(refreshToken.ExpiresAt) {
		return driver.TokenResponse{}, ErrRefreshTokenInvalid
	}

	// Отзываем старый refresh token
	_, err = s.db.ExecContext(ctx, "UPDATE driver_refresh_tokens SET is_revoked = true WHERE id = $1", refreshToken.ID)
	if err != nil {
		return driver.TokenResponse{}, err
	}

	// Получаем информацию о водителе
	var driverRecord driver.Driver
	err = s.db.GetContext(ctx, &driverRecord,
		`SELECT id, email, phone_number, first_name, last_name, license_number, 
		 license_expiry_date, vehicle_brand, vehicle_model, vehicle_number, vehicle_year, status, rating 
		 FROM drivers WHERE id = $1`, refreshToken.DriverID)
	if err != nil {
		return driver.TokenResponse{}, ErrDriverNotFound
	}

	driverInfo := driver.DriverInfo{
		ID:                driverRecord.ID,
		Email:             driverRecord.Email,
		PhoneNumber:       driverRecord.PhoneNumber,
		FirstName:         driverRecord.FirstName,
		LastName:          driverRecord.LastName,
		LicenseNumber:     driverRecord.LicenseNumber,
		LicenseExpiryDate: driverRecord.LicenseExpiryDate.Format("2006-01-02"),
		VehicleBrand:      driverRecord.VehicleBrand,
		VehicleModel:      driverRecord.VehicleModel,
		VehicleNumber:     driverRecord.VehicleNumber,
		VehicleYear:       driverRecord.VehicleYear,
		Status:            driverRecord.Status,
		Rating:            driverRecord.Rating,
	}

	// Генерируем новые токены
	return s.generateTokenPair(ctx, driverInfo)
}

func (s *service) ValidateToken(ctx context.Context, tokenString string) (driver.DriverInfo, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return driver.DriverInfo{}, ErrInvalidToken
	}

	if !token.Valid {
		return driver.DriverInfo{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return driver.DriverInfo{}, ErrInvalidToken
	}

	// Проверяем тип токена и роль
	tokenType, ok := claims["type"].(string)
	if !ok || tokenType != "access" {
		return driver.DriverInfo{}, ErrInvalidToken
	}

	role, ok := claims["role"].(string)
	if !ok || role != "driver" {
		return driver.DriverInfo{}, ErrInvalidToken
	}

	driverID := int64(claims["driver_id"].(float64))

	var driverRecord driver.Driver
	err = s.db.GetContext(ctx, &driverRecord,
		`SELECT id, email, phone_number, first_name, last_name, license_number, 
		 license_expiry_date, vehicle_brand, vehicle_model, vehicle_number, vehicle_year, status, rating 
		 FROM drivers WHERE id = $1`, driverID)
	if err != nil {
		return driver.DriverInfo{}, ErrDriverNotFound
	}

	return driver.DriverInfo{
		ID:                driverRecord.ID,
		Email:             driverRecord.Email,
		PhoneNumber:       driverRecord.PhoneNumber,
		FirstName:         driverRecord.FirstName,
		LastName:          driverRecord.LastName,
		LicenseNumber:     driverRecord.LicenseNumber,
		LicenseExpiryDate: driverRecord.LicenseExpiryDate.Format("2006-01-02"),
		VehicleBrand:      driverRecord.VehicleBrand,
		VehicleModel:      driverRecord.VehicleModel,
		VehicleNumber:     driverRecord.VehicleNumber,
		VehicleYear:       driverRecord.VehicleYear,
		Status:            driverRecord.Status,
		Rating:            driverRecord.Rating,
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	// Отзываем конкретный refresh token
	_, err := s.db.ExecContext(ctx, "UPDATE driver_refresh_tokens SET is_revoked = true WHERE token = $1", refreshToken)
	return err
}

func (s *service) LogoutAll(ctx context.Context, driverID int64) error {
	// Отзываем все refresh токены водителя
	_, err := s.db.ExecContext(ctx, "UPDATE driver_refresh_tokens SET is_revoked = true WHERE driver_id = $1", driverID)
	return err
}

func (s *service) generateTokenPair(ctx context.Context, driverInfo driver.DriverInfo) (driver.TokenResponse, error) {
	now := time.Now()

	// Создаем Access Token с ролью водителя
	accessClaims := jwt.MapClaims{
		"driver_id": driverInfo.ID,
		"email":     driverInfo.Email,
		"role":      "driver",
		"type":      "access",
		"exp":       now.Add(AccessTokenDuration).Unix(),
		"iat":       now.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return driver.TokenResponse{}, err
	}

	// Генерируем Refresh Token (случайная строка)
	refreshTokenBytes := make([]byte, 32)
	_, err = rand.Read(refreshTokenBytes)
	if err != nil {
		return driver.TokenResponse{}, err
	}
	refreshTokenString := hex.EncodeToString(refreshTokenBytes)

	// Сохраняем refresh token в БД
	refreshExpiresAt := now.Add(RefreshTokenDuration)
	query := `INSERT INTO driver_refresh_tokens (driver_id, token, expires_at, created_at, is_revoked)
			  VALUES ($1, $2, $3, $4, false)`

	_, err = s.db.ExecContext(ctx, query, driverInfo.ID, refreshTokenString, refreshExpiresAt, now)
	if err != nil {
		return driver.TokenResponse{}, err
	}

	return driver.TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenDuration.Seconds()),
		ExpiresAt:    now.Add(AccessTokenDuration),
		DriverInfo:   driverInfo,
	}, nil
}
