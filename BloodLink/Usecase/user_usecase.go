package Usecase

import (
	"context"
	"errors"
	"time"

	domain "bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"
	"bloodlink/Infrastructure"
	
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
}

type IProfileRepository interface {
	CreateProfile(ctx context.Context, profile *domain.UserProfile) error
	GetProfileByUserID(ctx context.Context, userID string) (*domain.UserProfile, error)
	UpdateProfile(ctx context.Context, profile *domain.UserProfile) error
}

type UserUseCaseBase struct {
	userRepo    IUserRepository
	profileRepo IProfileRepository
	auth        domainInterface.IAuthentication
	validation  domainInterface.IUserValidation
}

func NewUserUseCase(userRepo IUserRepository, profileRepo IProfileRepository, auth domainInterface.IAuthentication, validation domainInterface.IUserValidation) *UserUseCaseBase {
	return &UserUseCaseBase{
		userRepo:    userRepo,
		profileRepo: profileRepo,
		auth:        auth,
		validation:  validation,
	}
}

func (u *UserUseCaseBase) RegisterUser(ctx context.Context, user *domain.User) error {
	// Validate email
	if !u.validation.IsValidEmail(user.Email) {
		return errors.New("invalid email format")
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
	if user.Role == "" {
		user.Role = domain.RoleDonor
	}

	if user.Role == domain.RoleBloodBankAdmin {
		return errors.New("cannot register as Blood Bank Admin")
	}

	// 1. Save User to db
	if err := u.userRepo.CreateUser(ctx, user); err != nil {
		return err
	}

	// 2. Send OTP Email
	if err := Infrastructure.SendOTP(user.Email, user.OTP); err != nil {
		return errors.New("failed to send verification email: " + err.Error())
	}

	return nil
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

	// Create role-specific tables
	if user.Role == domain.RoleDonor {
		donor := &domain.Donor{
			DonorID: uuid.New().String(),
			UserID:  user.ID,
			Status:  "Available",
		}
		return u.userRepo.CreateDonor(ctx, donor)
	}

	return nil
}

func (u *UserUseCaseBase) GetProfile(ctx context.Context, userID string) (*domain.UserProfile, error) {
	return u.profileRepo.GetProfileByUserID(ctx, userID)
}

func (u *UserUseCaseBase) UpdateProfile(ctx context.Context, profile *domain.UserProfile) error {
	return u.profileRepo.UpdateProfile(ctx, profile)
}

func (u *UserUseCaseBase) DeleteUser(ctx context.Context, userID string) error {
	return u.userRepo.DeleteUser(ctx, userID)
}


func (u *UserUseCaseBase) Login(ctx context.Context, email, password string) (string, string, error) {
	// Hardcoded BloodBank Admin bypass
	if email == "admin@bloodlink.com" {
		if password != "Admin123!" {
			return "", "", errors.New("invalid credentials")
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
			return "", "", err
		}

		refreshToken, err := u.auth.GenerateToken(claims, domainInterface.RefreshToken)
		if err != nil {
			return "", "", err
		}

		return accessToken, refreshToken, nil
	}

	// Normal User flow
	// Get user by email
	user, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("invalid credentials") // Avoid "user not found" for security
	}

	// Compare password
	err = u.validation.ComparePassword(user.Password, password)
	if err != nil {
		return "", "", errors.New("invalid credentials")
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
		return "", "", err
	}

	// Generate Refresh Token
	refreshToken, err := u.auth.GenerateToken(claims, domainInterface.RefreshToken)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
