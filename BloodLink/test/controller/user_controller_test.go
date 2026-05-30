package controller_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	domain "bloodlink/Domain"
	"bloodlink/Delivery/controller"
	"bloodlink/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ─────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────

func newTestRouter(ctrl *controller.UserController) *gin.Engine {
	r := gin.New()
	r.POST("/auth/register", ctrl.RegisterUser)
	r.POST("/auth/register-donor", ctrl.RegisterDonor)
	r.POST("/auth/login", ctrl.HandleLogin)
	r.POST("/auth/verify-otp", ctrl.VerifyOTP)
	r.POST("/auth/forgot-password", ctrl.ForgotPassword)
	r.POST("/auth/reset-password", ctrl.ResetPassword)
	r.POST("/auth/refresh-token", ctrl.RefreshTokenHandler)

	// inject userID into context to simulate auth middleware
	authed := r.Group("/")
	authed.Use(func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Next()
	})
	authed.GET("/profile", ctrl.GetProfile)
	authed.GET("/profile/:id", ctrl.GetProfileByID)
	authed.PATCH("/profile", ctrl.UpdateProfile)
	authed.DELETE("/user", ctrl.DeleteUser)
	authed.POST("/logout", ctrl.Logout)
	authed.GET("/donors/filter", ctrl.GetDonors)

	return r
}

func postJSON(r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getJSON(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func patchJSON(r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func deleteReq(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ─────────────────────────────────────────────
// RegisterUser
// ─────────────────────────────────────────────

func TestRegisterUser_Handler_Success(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("RegisterUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	body := map[string]string{
		"full_name": "Test Collector",
		"email":     "collector@test.com",
		"phone":     "+251912345678",
		"password":  "StrongPass1!",
		"role":      "bloodcollector",
	}

	w := postJSON(r, "/auth/register", body)

	assert.Equal(t, http.StatusCreated, w.Code)
	uc.AssertExpectations(t)
}

func TestRegisterUser_Handler_UsecaseError(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("RegisterUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("email already registered"))

	body := map[string]string{
		"full_name": "Test",
		"email":     "dup@test.com",
		"phone":     "+251912345678",
		"password":  "StrongPass1!",
		"role":      "bloodcollector",
	}

	w := postJSON(r, "/auth/register", body)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRegisterUser_Handler_MissingFields(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	// Missing required fields
	body := map[string]string{"email": "test@test.com"}

	w := postJSON(r, "/auth/register", body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ─────────────────────────────────────────────
// Login
// ─────────────────────────────────────────────

func TestLogin_Handler_Success(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("Login", mock.Anything, "donor@test.com", "password123").
		Return("access-token", "refresh-token", domain.RoleDonor, "user-id", false, nil)

	body := map[string]string{
		"email":    "donor@test.com",
		"password": "password123",
	}

	w := postJSON(r, "/auth/login", body)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "access-token", resp["access_token"])
	assert.Equal(t, "refresh-token", resp["refresh_token"])
	assert.Equal(t, domain.RoleDonor, resp["role"])
}

func TestLogin_Handler_InvalidCredentials(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("Login", mock.Anything, "bad@test.com", "wrongpass").
		Return("", "", "", "", false, errors.New("invalid credentials"))

	body := map[string]string{
		"email":    "bad@test.com",
		"password": "wrongpass",
	}

	w := postJSON(r, "/auth/login", body)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_Handler_OTPNeeded_ForCollector(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("Login", mock.Anything, "collector@test.com", "pass").
		Return("at", "rt", domain.RoleBloodCollector, "uid", true, nil)

	body := map[string]string{
		"email":    "collector@test.com",
		"password": "pass",
	}

	w := postJSON(r, "/auth/login", body)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, true, resp["otp_needed"])
}

// ─────────────────────────────────────────────
// GetProfile
// ─────────────────────────────────────────────

func TestGetProfile_Handler_Success(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	profile := &domain.ProfileResponse{
		UserProfile: domain.UserProfile{
			ProfileID: "p1",
			UserID:    "test-user-id",
			FullName:  "Test User",
		},
	}

	uc.On("GetProfile", mock.Anything, "test-user-id").Return(profile, nil)

	w := getJSON(r, "/profile")

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Test User", resp["full_name"])
}

func TestGetProfile_Handler_NotFound(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("GetProfile", mock.Anything, "test-user-id").Return(nil, nil)

	w := getJSON(r, "/profile")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ─────────────────────────────────────────────
// GetProfileByID
// ─────────────────────────────────────────────

func TestGetProfileByID_Handler_Success(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	targetID := "hospital-admin-id"
	profile := &domain.ProfileResponse{
		UserProfile: domain.UserProfile{
			ProfileID: targetID,
			UserID:    targetID,
			FullName:  "Hospital Admin",
		},
	}

	uc.On("GetProfile", mock.Anything, targetID).Return(profile, nil)

	w := getJSON(r, "/profile/"+targetID)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Hospital Admin", resp["full_name"])
}

func TestGetProfileByID_Handler_NotFound(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("GetProfile", mock.Anything, "ghost-id").Return(nil, nil)

	w := getJSON(r, "/profile/ghost-id")

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ─────────────────────────────────────────────
// UpdateProfile
// ─────────────────────────────────────────────

func TestUpdateProfile_Handler_Success(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	existing := &domain.ProfileResponse{
		UserProfile: domain.UserProfile{
			ProfileID: "p1",
			UserID:    "test-user-id",
			FullName:  "Old Name",
		},
	}

	uc.On("GetProfile", mock.Anything, "test-user-id").Return(existing, nil)
	uc.On("UpdateProfile", mock.Anything, mock.AnythingOfType("*domain.UserProfile")).Return(nil)

	body := map[string]string{"full_name": "New Name"}

	w := patchJSON(r, "/profile", body)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ─────────────────────────────────────────────
// DeleteUser
// ─────────────────────────────────────────────

func TestDeleteUser_Handler_Success(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("DeleteUser", mock.Anything, "test-user-id").Return(nil)

	w := deleteReq(r, "/user")

	assert.Equal(t, http.StatusOK, w.Code)
}

// ─────────────────────────────────────────────
// ForgotPassword / ResetPassword
// ─────────────────────────────────────────────

func TestForgotPassword_Handler_Success(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("ForgotPassword", mock.Anything, "user@test.com").Return(nil)

	body := map[string]string{"email": "user@test.com"}
	w := postJSON(r, "/auth/forgot-password", body)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResetPassword_Handler_Success(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("ResetPassword", mock.Anything, "user@test.com", "123456", "NewPass1!").Return(nil)

	body := map[string]string{
		"email":        "user@test.com",
		"otp":          "123456",
		"new_password": "NewPass1!",
	}
	w := postJSON(r, "/auth/reset-password", body)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResetPassword_Handler_InvalidOTP(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("ResetPassword", mock.Anything, "user@test.com", "000000", "NewPass1!").Return(errors.New("invalid OTP"))

	body := map[string]string{
		"email":        "user@test.com",
		"otp":          "000000",
		"new_password": "NewPass1!",
	}
	w := postJSON(r, "/auth/reset-password", body)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ─────────────────────────────────────────────
// Logout
// ─────────────────────────────────────────────

func TestLogout_Handler_Success(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("Logout", mock.Anything, "test-user-id").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ─────────────────────────────────────────────
// RefreshToken
// ─────────────────────────────────────────────

func TestRefreshToken_Handler_Success(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("RefreshToken", mock.Anything, "old-refresh-token").
		Return("new-access-token", "new-refresh-token", nil)

	body := map[string]string{"refresh_token": "old-refresh-token"}
	w := postJSON(r, "/auth/refresh-token", body)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "new-access-token", resp["access_token"])
}

func TestRefreshToken_Handler_InvalidToken(t *testing.T) {
	uc := &mocks.MockUserUseCase{}
	ctrl := controller.NewUserController(uc)
	r := newTestRouter(ctrl)

	uc.On("RefreshToken", mock.Anything, "bad-token").
		Return("", "", errors.New("invalid or expired refresh token"))

	body := map[string]string{"refresh_token": "bad-token"}
	w := postJSON(r, "/auth/refresh-token", body)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
