package Usecase

import (
	"context"
	"errors"
	"log"
	"time"

	domain "bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"
	"bloodlink/Infrastructure"

	"fmt"

	"github.com/google/uuid"
)

// Instead of passing interface, we can pass our concrete repo for now, or update the interface.
// For clean architecture, we usually pass an interface.
type IUserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	ActivateUser(ctx context.Context, userID string) error
	CreateDonor(ctx context.Context, donor *domain.Donor) error
	DeleteUser(ctx context.Context, userID string) error
	FilterDonors(ctx context.Context, filter domain.DonorFilter) ([]domain.DonorResponse, error)
	SetOTP(ctx context.Context, email, otp string) error
	ResetPassword(ctx context.Context, email, hashedPassword string) error
	UpdateDonorStatus(ctx context.Context, donorID, status string) error
	GetUsersByRole(ctx context.Context, filter domain.UserFilter) ([]domain.UserResponse, error)
	UpdateRefreshToken(ctx context.Context, userID, refreshToken string) error
	GetDonorByUserID(ctx context.Context, userID string) (*domain.Donor, error)
	GetDonorsByBloodTypeAndAddress(ctx context.Context, bloodType, address string) ([]domain.DonorResponse, error)
	GetDonorsNearby(ctx context.Context, bloodType string, lat, lon, radiusKm float64) ([]domain.DonorResponse, error)
	GetUserByPhone(ctx context.Context, phone string) (*domain.User, error)
	GetEligibleDonors(ctx context.Context, query string) ([]domain.DonorResponse, error)
	GetDonorsBecomingEligibleToday(ctx context.Context) ([]domain.DonorResponse, error)
}

type IProfileRepository interface {
	CreateProfile(ctx context.Context, profile *domain.UserProfile) error
	GetProfileByUserID(ctx context.Context, userID string) (*domain.UserProfile, error)
	UpdateProfile(ctx context.Context, profile *domain.UserProfile) error
	GetAllProfiles(ctx context.Context, filter domain.ProfileFilter) ([]domain.UserProfile, error)
}

type UserUseCaseBase struct {
	userRepo     IUserRepository
	profileRepo  IProfileRepository
	donationRepo domainInterface.IDonationRepository
	auth         domainInterface.IAuthentication
	validation   domainInterface.IUserValidation
	notifUC      domainInterface.INotificationUsecase
}

func NewUserUseCase(
	userRepo IUserRepository,
	profileRepo IProfileRepository,
	donationRepo domainInterface.IDonationRepository,
	auth domainInterface.IAuthentication,
	validation domainInterface.IUserValidation,
	notifUC domainInterface.INotificationUsecase,
) *UserUseCaseBase {
	return &UserUseCaseBase{
		userRepo:     userRepo,
		profileRepo:  profileRepo,
		donationRepo: donationRepo,
		auth:         auth,
		validation:   validation,
		notifUC:      notifUC,
	}
}

func (u *UserUseCaseBase) RegisterUser(ctx context.Context, user *domain.User) error {
	// Validate email
	if !u.validation.IsValidEmail(user.Email) {
		return errors.New("invalid email format")
	}

	// Check if email already exists
	existingEmail, _ := u.userRepo.GetUserByEmail(ctx, user.Email)
	if existingEmail != nil {
		return errors.New("email already registered")
	}

	// Check if phone already exists
	if user.Phone != "" {
		existingPhone, _ := u.userRepo.GetUserByPhone(ctx, user.Phone)
		if existingPhone != nil {
			return errors.New("phone number already registered")
		}
	}

	// Validate password
	if !u.validation.IsStrongPassword(user.Password) {
		return errors.New("password is not strong enough")
	}

	// Hash password
	hashedPassword := u.validation.Hashpassword(user.Password)
	user.Password = hashedPassword

	// Set defaults
	user.ID = uuid.New().String()
	user.IsActive = false // Default to false for OTP verification
	user.CreatedAt = time.Now()
	user.OTP = Infrastructure.GenerateOTP()

	// Ensure role is valid
	if user.Role == domain.RoleDonor {
		return errors.New("please use the donor registration endpoint to register as a donor")
	}

	if user.Role == domain.RoleBloodBankAdmin {
		return errors.New("cannot register as Blood Bank Admin")
	}

	// 1. Save User to db
	if err := u.userRepo.CreateUser(ctx, user); err != nil {
		return err
	}

	// 2. Send OTP Email (Asynchronously to avoid timeouts on Render)
	go func(email, otp string) {
		if err := Infrastructure.SendOTP(email, otp); err != nil {
			log.Printf("[ERROR] Failed to send verification email to %s: %v", email, err)
		}
	}(user.Email, user.OTP)

	return nil
}

func (u *UserUseCaseBase) RegisterDonor(ctx context.Context, req *domain.RegisterDonorRequest) error {
	// 1. Validate
	if !u.validation.IsValidEmail(req.Email) {
		return errors.New("invalid email format")
	}
	if !u.validation.IsStrongPassword(req.Password) {
		return errors.New("password is not strong enough")
	}

	// Check if email already exists
	existingEmail, _ := u.userRepo.GetUserByEmail(ctx, req.Email)
	if existingEmail != nil {
		return errors.New("email already registered")
	}

	// Check if phone already exists
	if req.Phone != "" {
		existingPhone, _ := u.userRepo.GetUserByPhone(ctx, req.Phone)
		if existingPhone != nil {
			return errors.New("phone number already registered")
		}
	}

	// 2. Hash password
	hashedPassword := u.validation.Hashpassword(req.Password)

	// 3. Create User
	userID := uuid.New().String()
	user := &domain.User{
		ID:        userID,
		FullName:  req.FullName,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  hashedPassword,
		Role:      domain.RoleDonor,
		IsActive:  true, // No OTP verification
		CreatedAt: time.Now(),
	}

	if err := u.userRepo.CreateUser(ctx, user); err != nil {
		return err
	}

	// 4. Create Profile
	profile := &domain.UserProfile{
		ProfileID: uuid.New().String(),
		UserID:    userID,
		FullName:  req.FullName,
		Phone:     req.Phone,
		Address:   req.Address,
		Latitude:  &req.Latitude,
		Longitude: &req.Longitude,
	}

	if err := u.profileRepo.CreateProfile(ctx, profile); err != nil {
		return err
	}

	// 5. Create Donor
	donor := &domain.Donor{
		DonorID:       uuid.New().String(),
		UserID:        userID,
		DateOfBirth:   req.BirthDate,
		OverallStatus: "Pending",
	}

	return u.userRepo.CreateDonor(ctx, donor)
}

func (u *UserUseCaseBase) VerifyOTP(ctx context.Context, email, otp string) error {
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.OTP != otp {
		return errors.New("invalid OTP")
	}

	// Activate user
	if err := u.userRepo.ActivateUser(ctx, user.ID); err != nil {
		return err
	}

	// Now create the profile
	profile := &domain.UserProfile{
		ProfileID: uuid.New().String(),
		UserID:    user.ID,
		FullName:  user.FullName,
		Phone:     user.Phone,
	}

	if err := u.profileRepo.CreateProfile(ctx, profile); err != nil {
		return err
	}

	return nil
}

func (u *UserUseCaseBase) GetProfile(ctx context.Context, userID string) (*domain.ProfileResponse, error) {
	profile, err := u.profileRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, nil
	}

	response := &domain.ProfileResponse{
		UserProfile: *profile,
	}

	// Check if user is a donor
	donor, err := u.userRepo.GetDonorByUserID(ctx, userID)
	if err == nil && donor != nil {
		// It's a donor, calculate eligibility
		eligibility := u.CalculateEligibility(ctx, donor.DonorID, donor.OverallStatus)
		response.DonorInfo = donor
		response.Eligibility = eligibility
	}

	return response, nil
}

func (u *UserUseCaseBase) CalculateEligibility(ctx context.Context, donorID string, overallStatus string) *domain.DonorEligibility {
	eligibility := &domain.DonorEligibility{
		IsEligible:        false,
		EligibilityStatus: "Not Eligible",
	}

	// 1. Check Permanent Deferral
	if overallStatus == "PERMANENTLY_DEFERRED" {
		eligibility.EligibilityMessage = "Not Eligible — You are permanently deferred from donating blood and cannot donate again. Thank you for your willingness and your past contribution to saving lives."
		return eligibility
	}

	// 2. Check 3-month rule
	lastDonation, err := u.donationRepo.GetLastDonationByDonor(donorID)
	if err != nil || lastDonation == nil {
		// Never donated before
		eligibility.IsEligible = true
		eligibility.EligibilityStatus = "Eligible"
		eligibility.EligibilityMessage = "Eligible — You can come and donate at any time. Thank you for your willingness to save lives."
		return eligibility
	}

	// Calculate days since last donation
	daysSince := int(time.Since(lastDonation.CollectionDate).Hours() / 24)
	remainingDays := 90 - daysSince

	if daysSince >= 90 {
		// More than 3 months passed
		eligibility.IsEligible = true
		eligibility.EligibilityStatus = "Eligible"
		eligibility.EligibilityMessage = "Eligible — You can come and donate at any time. Thank you for your willingness to save lives."
	} else {
		// Within 3 months
		eligibility.CountdownDays = remainingDays
		if overallStatus == "Pending" {
			eligibility.EligibilityMessage = fmt.Sprintf("Not Eligible — Your overall lab status is pending. Please wait %d more days before your next donation.", remainingDays)
		} else {
			eligibility.EligibilityMessage = fmt.Sprintf("Not Eligible — You donated recently. Please wait %d more days before donating again.", remainingDays)
		}
	}

	return eligibility
}

func (u *UserUseCaseBase) GetAllProfiles(ctx context.Context, filter domain.ProfileFilter) ([]domain.UserProfile, error) {
	return u.profileRepo.GetAllProfiles(ctx, filter)
}

func (u *UserUseCaseBase) UpdateProfile(ctx context.Context, profile *domain.UserProfile) error {
	return u.profileRepo.UpdateProfile(ctx, profile)
}

func (u *UserUseCaseBase) DeleteUser(ctx context.Context, userID string) error {
	return u.userRepo.DeleteUser(ctx, userID)
}

func (u *UserUseCaseBase) FilterDonors(ctx context.Context, filter domain.DonorFilter) ([]domain.DonorResponse, error) {
	return u.userRepo.FilterDonors(ctx, filter)
}

func (u *UserUseCaseBase) UpdateDonorStatus(ctx context.Context, donorID, status string) error {
	validStatuses := map[string]bool{
		"Pending":              true,
		"CLEARED":              true,
		"TEMPORARILY_DEFERRED": true,
		"PERMANENTLY_DEFERRED": true,
	}
	if !validStatuses[status] {
		return errors.New("invalid status: must be Pending, CLEARED, TEMPORARILY_DEFERRED, or PERMANENTLY_DEFERRED")
	}
	return u.userRepo.UpdateDonorStatus(ctx, donorID, status)
}

func (u *UserUseCaseBase) GetUsersByRole(ctx context.Context, filter domain.UserFilter) ([]domain.UserResponse, error) {
	return u.userRepo.GetUsersByRole(ctx, filter)
}

func (u *UserUseCaseBase) ForgotPassword(ctx context.Context, email string) error {
	// Verify user exists
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Generate and store OTP
	otp := Infrastructure.GenerateOTP()
	if err := u.userRepo.SetOTP(ctx, email, otp); err != nil {
		return err
	}

	// Send OTP email asynchronously
	go func() {
		if err := Infrastructure.SendPasswordResetOTP(email, otp); err != nil {
			log.Printf("[ERROR] Failed to send password reset email to %s: %v", email, err)
		}
	}()

	return nil
}

func (u *UserUseCaseBase) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	// Get user and verify OTP
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}
	if user.OTP != otp {
		return errors.New("invalid OTP")
	}

	// Validate and hash new password
	if !u.validation.IsStrongPassword(newPassword) {
		return errors.New("password is not strong enough")
	}
	hashedPassword := u.validation.Hashpassword(newPassword)

	return u.userRepo.ResetPassword(ctx, email, hashedPassword)
}

func (u *UserUseCaseBase) Login(ctx context.Context, email, password string) (string, string, string, string, error) {
	// Hardcoded BloodBank Admin bypass
	if email == "admin@bloodlink.com" {
		if password != "Admin123!" {
			return "", "", "", "", errors.New("invalid credentials")
		}

		// Create claims for the hardcoded admin
		claims := &domain.UserClaims{
			UserID:      "00000000-0000-0000-0000-000000000000",
			Email:       "admin@bloodlink.com",
			AccountType: domain.RoleBloodBankAdmin,
			IsVerified:  true,
		}

		accessToken, err := u.auth.GenerateToken(claims, domainInterface.AccessToken)
		if err != nil {
			return "", "", "", "", err
		}

		refreshToken, err := u.auth.GenerateToken(claims, domainInterface.RefreshToken)
		if err != nil {
			return "", "", "", "", err
		}

		// Save refresh token to DB (bypass for hardcoded admin usually, but let's keep it consistent if possible)
		// Actually, admin@bloodlink.com doesn't exist in DB, so u.userRepo.UpdateRefreshToken will fail.
		// Let's only do it for normal users.

		return accessToken, refreshToken, "bloodbankadmin", claims.UserID, nil
	}

	// Normal User flow
	// Get user by email
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", "", "", err
	}
	if user == nil {
		return "", "", "", "", errors.New("invalid credentials") // Avoid "user not found" for security
	}

	// Compare password
	err = u.validation.ComparePassword(user.Password, password)
	if err != nil {
		return "", "", "", "", errors.New("invalid credentials")
	}

	// Create claims
	claims := &domain.UserClaims{
		UserID:      user.ID,
		Email:       user.Email,
		AccountType: user.Role,
		IsVerified:  user.IsActive,
	}

	// Generate Access Token
	accessToken, err := u.auth.GenerateToken(claims, domainInterface.AccessToken)
	if err != nil {
		return "", "", "", "", err
	}

	// Generate Refresh Token
	refreshToken, err := u.auth.GenerateToken(claims, domainInterface.RefreshToken)
	if err != nil {
		return "", "", "", "", err
	}

	// Save Refresh Token to DB
	if err := u.userRepo.UpdateRefreshToken(ctx, user.ID, refreshToken); err != nil {
		return "", "", "", "", err
	}

	return accessToken, refreshToken, user.Role, user.ID, nil
}

func (u *UserUseCaseBase) RefreshToken(ctx context.Context, refreshTokenStr string) (string, string, error) {
	// Parse and validate the refresh token
	claims, err := u.auth.ParseTokenToClaim(refreshTokenStr)
	if err != nil {
		return "", "", errors.New("invalid or expired refresh token")
	}

	// Double check the token type
	if claims.TokenType != domainInterface.RefreshToken {
		return "", "", errors.New("invalid token type")
	}

	// For normal users, verify they still exist and are active
	// Hardcoded Admin bypass check
	if claims.Email != "admin@bloodlink.com" {
		user, err := u.userRepo.GetUserByEmail(ctx, claims.Email)
		if err != nil || user == nil {
			return "", "", errors.New("user no longer exists")
		}

		// REVOCATION CHECK: Verify stored token matches provided token
		if user.RefreshToken != refreshTokenStr {
			return "", "", errors.New("refresh token has been revoked or session expired")
		}

		// Refresh the claims with current user data
		claims.AccountType = user.Role
		claims.IsVerified = user.IsActive
	}

	// Generate new tokens
	newAccessToken, err := u.auth.GenerateToken(claims, domainInterface.AccessToken)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := u.auth.GenerateToken(claims, domainInterface.RefreshToken)
	if err != nil {
		return "", "", err
	}

	// Update the stored refresh token (Token Rotation)
	if claims.Email != "admin@bloodlink.com" {
		if err := u.userRepo.UpdateRefreshToken(ctx, claims.UserID, newRefreshToken); err != nil {
			return "", "", err
		}
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *UserUseCaseBase) Logout(ctx context.Context, userID string) error {
	// Simple revocation: clear the refresh token in the database
	return u.userRepo.UpdateRefreshToken(ctx, userID, "")
}

func (u *UserUseCaseBase) GetDonorIDByUserID(ctx context.Context, userID string) (string, error) {
	donor, err := u.userRepo.GetDonorByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	return donor.DonorID, nil
}

func (u *UserUseCaseBase) GetEligibleDonors(ctx context.Context, query string) ([]domain.DonorResponse, error) {
	return u.userRepo.GetEligibleDonors(ctx, query)
}

func (u *UserUseCaseBase) NotifyEligibleDonors(ctx context.Context) error {
	donors, err := u.userRepo.GetDonorsBecomingEligibleToday(ctx)
	if err != nil {
		return err
	}

	for _, d := range donors {
		subject := "You are Eligible to Donate Again!"
		content := "Eligible — You can come and donate at any time. Thank you for your willingness to save lives."
		go u.notifUC.SendNotification(d.UserID, "ELIGIBILITY", subject, content)
	}
	return nil
}
func (u *UserUseCaseBase) UpdateLocation(ctx context.Context, userID string, lat, lon float64) error {
	profile, err := u.profileRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if profile == nil {
		return errors.New("profile not found")
	}

	profile.Latitude = &lat
	profile.Longitude = &lon

	return u.profileRepo.UpdateProfile(ctx, profile)
}
