package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "bloodlink/Domain"
	"bloodlink/Usecase"
	"bloodlink/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────

func buildUserUsecase(
	userRepo *mocks.MockUserRepository,
	profileRepo *mocks.MockProfileRepository,
	donationRepo *mocks.MockDonationRepository,
	auth *mocks.MockAuthentication,
	validation *mocks.MockUserValidation,
	notif *mocks.MockNotificationUsecase,
	hospitalRepo *mocks.MockHospitalRepository,
) *Usecase.UserUseCaseBase {
	return Usecase.NewUserUseCase(
		userRepo,
		profileRepo,
		donationRepo,
		auth,
		validation,
		notif,
		hospitalRepo,
	)
}

// ─────────────────────────────────────────────
// RegisterUser
// ─────────────────────────────────────────────

func TestRegisterUser_Success(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	user := &domain.User{
		Email:    "collector@test.com",
		Password: "StrongPass1!",
		Role:     domain.RoleBloodCollector,
		FullName: "Test Collector",
		Phone:    "+251912345678",
	}

	validation.On("IsValidPhone", user.Phone).Return(true)
	validation.On("IsValidEmail", user.Email).Return(true)
	validation.On("IsStrongPassword", user.Password).Return(true)
	validation.On("Hashpassword", user.Password).Return("hashed_password")
	userRepo.On("GetUserByEmail", mock.Anything, user.Email).Return(nil, nil)
	userRepo.On("GetUserByPhone", mock.Anything, user.Phone).Return(nil, nil)
	userRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*Domain.User")).Return(nil)

	err := uc.RegisterUser(context.Background(), user)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	validation.AssertExpectations(t)
}

func TestRegisterUser_DuplicateEmail(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	user := &domain.User{
		Email:    "existing@test.com",
		Password: "StrongPass1!",
		Role:     domain.RoleBloodCollector,
		Phone:    "+251912345678",
	}

	validation.On("IsValidPhone", user.Phone).Return(true)
	validation.On("IsValidEmail", user.Email).Return(true)
	userRepo.On("GetUserByEmail", mock.Anything, user.Email).Return(&domain.User{Email: user.Email}, nil)

	err := uc.RegisterUser(context.Background(), user)

	assert.EqualError(t, err, "email already registered")
}

func TestRegisterUser_InvalidPhone(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	user := &domain.User{
		Email:    "test@test.com",
		Password: "StrongPass1!",
		Role:     domain.RoleBloodCollector,
		Phone:    "0912345678", // missing +251 prefix
	}

	validation.On("IsValidPhone", user.Phone).Return(false)

	err := uc.RegisterUser(context.Background(), user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "phone number must follow")
}

func TestRegisterUser_DonorRoleBlocked(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	user := &domain.User{
		Email:    "donor@test.com",
		Password: "StrongPass1!",
		Role:     domain.RoleDonor,
		Phone:    "+251912345678",
	}

	validation.On("IsValidPhone", user.Phone).Return(true)
	validation.On("IsValidEmail", user.Email).Return(true)
	userRepo.On("GetUserByEmail", mock.Anything, user.Email).Return(nil, nil)
	userRepo.On("GetUserByPhone", mock.Anything, user.Phone).Return(nil, nil)
	validation.On("IsStrongPassword", user.Password).Return(true)
	validation.On("Hashpassword", user.Password).Return("hashed")

	err := uc.RegisterUser(context.Background(), user)

	assert.EqualError(t, err, "please use the donor registration endpoint to register as a donor")
}

// ─────────────────────────────────────────────
// GetProfile
// ─────────────────────────────────────────────

func TestGetProfile_DonorWithProfile(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	userID := "donor-user-id"
	profile := &domain.UserProfile{
		ProfileID: "profile-id",
		UserID:    userID,
		FullName:  "John Donor",
		Phone:     "+251911111111",
	}

	profileRepo.On("GetProfileByUserID", mock.Anything, userID).Return(profile, nil)
	userRepo.On("GetDonorByUserID", mock.Anything, userID).Return(nil, errors.New("not a donor"))

	result, err := uc.GetProfile(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "John Donor", result.FullName)
	profileRepo.AssertExpectations(t)
}

func TestGetProfile_HospitalAdmin_NoProfileRow_FallsBackToUser(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	userID := "hospital-admin-user-id"
	user := &domain.User{
		ID:       userID,
		FullName: "Hospital Admin",
		Phone:    "+251922222222",
		Role:     domain.RoleHospitalAdmin,
	}

	// No profile row in user_profiles
	profileRepo.On("GetProfileByUserID", mock.Anything, userID).Return(nil, nil)
	// Falls back to users table
	userRepo.On("GetUserByID", mock.Anything, userID).Return(user, nil)
	userRepo.On("GetDonorByUserID", mock.Anything, userID).Return(nil, errors.New("not a donor"))

	result, err := uc.GetProfile(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Hospital Admin", result.FullName)
	assert.Equal(t, userID, result.UserID)
}

func TestGetProfile_HardcodedAdmin_ReturnsStaticProfile(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	// The hardcoded admin ID — no DB row exists for this
	adminID := "00000000-0000-0000-0000-000000000000"

	result, err := uc.GetProfile(context.Background(), adminID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Blood Bank Admin", result.FullName)
	assert.Equal(t, adminID, result.UserID)

	// No DB calls should be made for the hardcoded admin
	profileRepo.AssertNotCalled(t, "GetProfileByUserID")
	userRepo.AssertNotCalled(t, "GetUserByID")
}

func TestGetProfile_UserNotFound_ReturnsNil(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	userID := "ghost-user-id"

	profileRepo.On("GetProfileByUserID", mock.Anything, userID).Return(nil, nil)
	userRepo.On("GetUserByID", mock.Anything, userID).Return(nil, nil)

	result, err := uc.GetProfile(context.Background(), userID)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

// ─────────────────────────────────────────────
// UpdateProfile — upsert behaviour
// ─────────────────────────────────────────────

func TestUpdateProfile_ExistingRow_CallsUpdate(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	profile := &domain.UserProfile{
		UserID:   "donor-user-id",
		FullName: "Updated Name",
	}
	existing := &domain.UserProfile{ProfileID: "existing-profile-id", UserID: profile.UserID}

	profileRepo.On("GetProfileByUserID", mock.Anything, profile.UserID).Return(existing, nil)
	profileRepo.On("UpdateProfile", mock.Anything, profile).Return(nil)

	err := uc.UpdateProfile(context.Background(), profile)

	assert.NoError(t, err)
	profileRepo.AssertCalled(t, "UpdateProfile", mock.Anything, profile)
	profileRepo.AssertNotCalled(t, "CreateProfile")
}

func TestUpdateProfile_NoRow_CallsCreate(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	profile := &domain.UserProfile{
		UserID:   "hospital-admin-id",
		FullName: "Hospital Admin Updated",
	}

	// No existing row
	profileRepo.On("GetProfileByUserID", mock.Anything, profile.UserID).Return(nil, nil)
	profileRepo.On("CreateProfile", mock.Anything, mock.AnythingOfType("*Domain.UserProfile")).Return(nil)

	err := uc.UpdateProfile(context.Background(), profile)

	assert.NoError(t, err)
	profileRepo.AssertCalled(t, "CreateProfile", mock.Anything, mock.AnythingOfType("*Domain.UserProfile"))
	profileRepo.AssertNotCalled(t, "UpdateProfile")
}

// ─────────────────────────────────────────────
// Login
// ─────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	user := &domain.User{
		ID:       "user-id",
		Email:    "donor@test.com",
		Password: "hashed_password",
		Role:     domain.RoleDonor,
		IsActive: true,
	}

	userRepo.On("GetUserByEmail", mock.Anything, "donor@test.com").Return(user, nil)
	validation.On("ComparePassword", "hashed_password", "plaintext").Return(nil)
	auth.On("GenerateToken", mock.AnythingOfType("*Domain.UserClaims"), "access_token").Return("access-token-value", nil)
	auth.On("GenerateToken", mock.AnythingOfType("*Domain.UserClaims"), "refresh_token").Return("refresh-token-value", nil)
	userRepo.On("UpdateRefreshToken", mock.Anything, user.ID, "refresh-token-value").Return(nil)

	at, rt, role, uid, otpNeeded, err := uc.Login(context.Background(), "donor@test.com", "plaintext")

	assert.NoError(t, err)
	assert.Equal(t, "access-token-value", at)
	assert.Equal(t, "refresh-token-value", rt)
	assert.Equal(t, domain.RoleDonor, role)
	assert.Equal(t, user.ID, uid)
	assert.False(t, otpNeeded)
}

func TestLogin_HardcodedAdmin_Bypass(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	auth.On("GenerateToken", mock.AnythingOfType("*Domain.UserClaims"), "access_token").Return("admin-access-token", nil)
	auth.On("GenerateToken", mock.AnythingOfType("*Domain.UserClaims"), "refresh_token").Return("admin-refresh-token", nil)

	at, rt, role, uid, _, err := uc.Login(context.Background(), "admin@bloodlink.com", "Admin123!")

	assert.NoError(t, err)
	assert.Equal(t, "admin-access-token", at)
	assert.Equal(t, "admin-refresh-token", rt)
	assert.Equal(t, "bloodbankadmin", role)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", uid)

	// No DB calls for hardcoded admin
	userRepo.AssertNotCalled(t, "GetUserByEmail")
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	user := &domain.User{
		ID:       "user-id",
		Email:    "donor@test.com",
		Password: "hashed_password",
		Role:     domain.RoleDonor,
		IsActive: true,
	}

	userRepo.On("GetUserByEmail", mock.Anything, "donor@test.com").Return(user, nil)
	validation.On("ComparePassword", "hashed_password", "wrongpassword").Return(errors.New("mismatch"))

	_, _, _, _, _, err := uc.Login(context.Background(), "donor@test.com", "wrongpassword")

	assert.EqualError(t, err, "invalid credentials")
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	userRepo.On("GetUserByEmail", mock.Anything, "ghost@test.com").Return(nil, nil)
	hospitalRepo.On("IsAdminEmailPending", "ghost@test.com").Return(false, nil)

	_, _, _, _, _, err := uc.Login(context.Background(), "ghost@test.com", "anypassword")

	assert.EqualError(t, err, "invalid credentials")
}

// ─────────────────────────────────────────────
// VerifyOTP
// ─────────────────────────────────────────────

func TestVerifyOTP_Success(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	expiry := time.Now().Add(5 * time.Minute)
	user := &domain.User{
		ID:           "user-id",
		Email:        "labtech@test.com",
		OTP:          "123456",
		OTPExpiresAt: &expiry,
		IsActive:     false,
	}

	userRepo.On("GetUserByEmail", mock.Anything, "labtech@test.com").Return(user, nil)
	userRepo.On("ActivateUser", mock.Anything, user.ID).Return(nil)
	profileRepo.On("CreateProfile", mock.Anything, mock.AnythingOfType("*Domain.UserProfile")).Return(nil)

	err := uc.VerifyOTP(context.Background(), "labtech@test.com", "123456")

	assert.NoError(t, err)
	userRepo.AssertCalled(t, "ActivateUser", mock.Anything, user.ID)
	profileRepo.AssertCalled(t, "CreateProfile", mock.Anything, mock.AnythingOfType("*Domain.UserProfile"))
}

func TestVerifyOTP_InvalidOTP(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	expiry := time.Now().Add(5 * time.Minute)
	user := &domain.User{
		ID:           "user-id",
		Email:        "labtech@test.com",
		OTP:          "123456",
		OTPExpiresAt: &expiry,
	}

	userRepo.On("GetUserByEmail", mock.Anything, "labtech@test.com").Return(user, nil)

	err := uc.VerifyOTP(context.Background(), "labtech@test.com", "999999")

	assert.EqualError(t, err, "invalid OTP")
}

func TestVerifyOTP_Expired(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	expiry := time.Now().Add(-1 * time.Minute) // already expired
	user := &domain.User{
		ID:           "user-id",
		Email:        "labtech@test.com",
		OTP:          "123456",
		OTPExpiresAt: &expiry,
	}

	userRepo.On("GetUserByEmail", mock.Anything, "labtech@test.com").Return(user, nil)

	err := uc.VerifyOTP(context.Background(), "labtech@test.com", "123456")

	assert.EqualError(t, err, "OTP has expired")
}

// ─────────────────────────────────────────────
// UpdateDonorStatus
// ─────────────────────────────────────────────

func TestUpdateDonorStatus_ValidStatus(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	userRepo.On("UpdateDonorStatus", mock.Anything, "donor-id", "CLEARED").Return(nil)

	err := uc.UpdateDonorStatus(context.Background(), "donor-id", "CLEARED")

	assert.NoError(t, err)
}

func TestUpdateDonorStatus_InvalidStatus(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	err := uc.UpdateDonorStatus(context.Background(), "donor-id", "INVALID_STATUS")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

// ─────────────────────────────────────────────
// Logout
// ─────────────────────────────────────────────

func TestLogout_ClearsRefreshToken(t *testing.T) {
	userRepo := &mocks.MockUserRepository{}
	profileRepo := &mocks.MockProfileRepository{}
	donationRepo := &mocks.MockDonationRepository{}
	auth := &mocks.MockAuthentication{}
	validation := &mocks.MockUserValidation{}
	notif := &mocks.MockNotificationUsecase{}
	hospitalRepo := &mocks.MockHospitalRepository{}

	uc := buildUserUsecase(userRepo, profileRepo, donationRepo, auth, validation, notif, hospitalRepo)

	userRepo.On("UpdateRefreshToken", mock.Anything, "user-id", "").Return(nil)

	err := uc.Logout(context.Background(), "user-id")

	assert.NoError(t, err)
	userRepo.AssertCalled(t, "UpdateRefreshToken", mock.Anything, "user-id", "")
}
