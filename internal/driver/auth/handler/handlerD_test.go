package handler

import (
	"FixDrive/models/driver"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock service
type MockDriverService struct {
	mock.Mock
}

func (m *MockDriverService) Register(ctx context.Context, req driver.RegisterRequest) (driver.TokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(driver.TokenResponse), args.Error(1)
}

func (m *MockDriverService) Login(ctx context.Context, req driver.LoginRequest) (driver.TokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(driver.TokenResponse), args.Error(1)
}

func (m *MockDriverService) RefreshToken(ctx context.Context, req driver.RefreshTokenRequest) (driver.TokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(driver.TokenResponse), args.Error(1)
}

func (m *MockDriverService) ValidateToken(ctx context.Context, token string) (driver.DriverInfo, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(driver.DriverInfo), args.Error(1)
}

func (m *MockDriverService) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockDriverService) LogoutAll(ctx context.Context, driverID int64) error {
	args := m.Called(ctx, driverID)
	return args.Error(0)
}

func (m *MockDriverService) GetAllDrivers(ctx context.Context) ([]driver.DriverInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]driver.DriverInfo), args.Error(1)
}

func TestDriverHandler_Register_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewHandler(mockService)

	registerReq := driver.RegisterRequest{
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

	expectedResp := driver.TokenResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    1800,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		DriverInfo: driver.DriverInfo{
			ID:            1,
			Email:         "driver@example.com",
			FirstName:     "Mike",
			LastName:      "Driver",
			LicenseNumber: "DL123456789",
			Status:        "pending",
		},
	}

	mockService.On("Register", mock.Anything, registerReq).
		Return(expectedResp, nil)

	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response driver.TokenResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Equal(t, expectedResp.AccessToken, response.AccessToken)
	assert.Equal(t, expectedResp.DriverInfo.Email, response.DriverInfo.Email)
	assert.Equal(t, expectedResp.DriverInfo.LicenseNumber, response.DriverInfo.LicenseNumber)

	mockService.AssertExpectations(t)
}

func TestDriverHandler_Login_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewHandler(mockService)

	loginReq := driver.LoginRequest{
		Email:    "driver@example.com",
		Password: "password123",
	}

	expectedResp := driver.TokenResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "Bearer",
		DriverInfo: driver.DriverInfo{
			ID:    1,
			Email: "driver@example.com",
		},
	}

	mockService.On("Login", mock.Anything, loginReq).
		Return(expectedResp, nil)

	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response driver.TokenResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Equal(t, expectedResp.AccessToken, response.AccessToken)

	mockService.AssertExpectations(t)
}

func TestDriverHandler_GetProfile_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewHandler(mockService)

	expectedDriver := driver.DriverInfo{
		ID:            1,
		Email:         "driver@example.com",
		FirstName:     "Mike",
		LastName:      "Driver",
		LicenseNumber: "DL123456789",
		Status:        "active",
	}

	mockService.On("ValidateToken", mock.Anything, "valid_token").
		Return(expectedDriver, nil)

	req := httptest.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer valid_token")

	rr := httptest.NewRecorder()
	handler.GetProfile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response driver.DriverInfo
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Equal(t, expectedDriver.Email, response.Email)
	assert.Equal(t, expectedDriver.LicenseNumber, response.LicenseNumber)

	mockService.AssertExpectations(t)
}

func TestDriverHandler_GetProfile_MissingToken(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest("GET", "/me", nil)

	rr := httptest.NewRecorder()
	handler.GetProfile(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Missing token")
}
