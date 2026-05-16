package Domain

import (
	domain "bloodlink/Domain"
	"context"
)

type IUserUseCase interface {
	RegisterUser(ctx context.Context, user *domain.User) error
	Login(ctx context.Context, email, password string) (string, string, string, string, error)
	RegisterDonor(ctx context.Context, req *domain.RegisterDonorRequest) error
	VerifyOTP(ctx context.Context, email, otp string) error
	GetProfile(ctx context.Context, userID string) (*domain.ProfileResponse, error)
	UpdateProfile(ctx context.Context, profile *domain.UserProfile) error
	DeleteUser(ctx context.Context, userID string) error
	FilterDonors(ctx context.Context, filter domain.DonorFilter) (*domain.AllDonorsResponse, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, otp, newPassword string) error
	UpdateDonorStatus(ctx context.Context, donorID, status string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	Logout(ctx context.Context, userID string) error
	GetUsersByRole(ctx context.Context, filter domain.UserFilter) ([]domain.UserResponse, error)
	GetAllProfiles(ctx context.Context, filter domain.ProfileFilter) ([]domain.UserProfile, error)
	GetDonorIDByUserID(ctx context.Context, userID string) (string, error)
	GetEligibleDonors(ctx context.Context, query string) (*domain.EligibleDonorsResponse, error)
	GetEligibleDonorByID(ctx context.Context, id string) (*domain.DonorResponse, error)
	NotifyEligibleDonors(ctx context.Context) error
	UpdateLocation(ctx context.Context, userID string, lat, lon float64) error
}
