package handler

import (
	"FixDrive/models/client"
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
type MockClientService struct {
	mock.Mock
}

func (m *MockClientService) Register(ctx context.Context, req client.RegisterRequest) (client.TokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(client.TokenResponse), args.Error(1)
}

func (m *MockClientService) Login(ctx context.Context, req client.LoginRequest) (client.TokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(client.TokenResponse), args.Error(1)
}

func (m *MockClientService) RefreshToken(ctx context.Context, req client.RefreshTokenRequest) (client.TokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(client.TokenResponse), args.Error(1)
}

func (m *MockClientService) ValidateToken(ctx context.Context, token string) (client.UserInfo, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(client.UserInfo), args.Error(1)
}

func (m *MockClientService) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockClientService) LogoutAll(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockClientService) GetAllUsers(ctx context.Context) ([]client.UserInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]client.UserInfo), args.Error(1)
}

func TestHandler_Register_Success(t *testing.T) {
	mockService := new(MockClientService)
	handler := NewHandler(mockService)

	registerReq := client.RegisterRequest{
		Email:       "test@example.com",
		Password:    "password123",
		PhoneNumber: "+1234567890",
		FirstName:   "John",
		LastName:    "Doe",
	}

	expectedResp := client.TokenResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    1800,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
		UserInfo: client.UserInfo{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
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

	var response client.TokenResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Equal(t, expectedResp.AccessToken, response.AccessToken)
	assert.Equal(t, expectedResp.UserInfo.Email, response.UserInfo.Email)

	mockService.AssertExpectations(t)
}

func TestHandler_Register_InvalidJSON(t *testing.T) {
	mockService := new(MockClientService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Invalid JSON")
}

func TestHandler_Login_Success(t *testing.T) {
	mockService := new(MockClientService)
	handler := NewHandler(mockService)

	loginReq := client.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	expectedResp := client.TokenResponse{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		TokenType:    "Bearer",
		UserInfo: client.UserInfo{
			ID:    1,
			Email: "test@example.com",
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

	var response client.TokenResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Equal(t, expectedResp.AccessToken, response.AccessToken)

	mockService.AssertExpectations(t)
}

func TestHandler_GetProfile_Success(t *testing.T) {
	mockService := new(MockClientService)
	handler := NewHandler(mockService)

	expectedUser := client.UserInfo{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	mockService.On("ValidateToken", mock.Anything, "valid_token").
		Return(expectedUser, nil)

	req := httptest.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", "Bearer valid_token")

	rr := httptest.NewRecorder()
	handler.GetProfile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response client.UserInfo
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Equal(t, expectedUser.Email, response.Email)

	mockService.AssertExpectations(t)
}

func TestHandler_GetProfile_MissingToken(t *testing.T) {
	mockService := new(MockClientService)
	handler := NewHandler(mockService)

	req := httptest.NewRequest("GET", "/me", nil)

	rr := httptest.NewRecorder()
	handler.GetProfile(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "Missing token")
}

func TestHandler_RefreshToken_Success(t *testing.T) {
	mockService := new(MockClientService)
	handler := NewHandler(mockService)

	refreshReq := client.RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	expectedResp := client.TokenResponse{
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    1800,
		UserInfo: client.UserInfo{
			ID:    1,
			Email: "test@example.com",
		},
	}

	mockService.On("RefreshToken", mock.Anything, refreshReq).
		Return(expectedResp, nil)

	body, _ := json.Marshal(refreshReq)
	req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.RefreshToken(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response client.TokenResponse
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Equal(t, expectedResp.AccessToken, response.AccessToken)
	assert.Equal(t, expectedResp.RefreshToken, response.RefreshToken)

	mockService.AssertExpectations(t)
}

func TestHandler_Logout_Success(t *testing.T) {
	mockService := new(MockClientService)
	handler := NewHandler(mockService)

	logoutReq := client.RefreshTokenRequest{
		RefreshToken: "token_to_revoke",
	}

	mockService.On("Logout", mock.Anything, "token_to_revoke").
		Return(nil)

	body, _ := json.Marshal(logoutReq)
	req := httptest.NewRequest("POST", "/logout", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Logout(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Equal(t, "Logged out successfully", response["message"])

	mockService.AssertExpectations(t)
}

func TestHandler_LogoutAll_Success(t *testing.T) {
	mockService := new(MockClientService)
	handler := NewHandler(mockService)

	userInfo := client.UserInfo{
		ID:    1,
		Email: "test@example.com",
	}

	mockService.On("ValidateToken", mock.Anything, "valid_token").
		Return(userInfo, nil)
	mockService.On("LogoutAll", mock.Anything, int64(1)).
		Return(nil)

	req := httptest.NewRequest("POST", "/logout-all", nil)
	req.Header.Set("Authorization", "Bearer valid_token")

	rr := httptest.NewRecorder()
	handler.LogoutAll(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	json.Unmarshal(rr.Body.Bytes(), &response)

	assert.Equal(t, "Logged out from all devices", response["message"])

	mockService.AssertExpectations(t)
}
