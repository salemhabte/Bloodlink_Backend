package Domain

import (
	domain "bloodlink/Domain"
	"context"
)

type IUserUseCase interface {
	RegisterUser(ctx context.Context, user *domain.User) error
	Login(ctx context.Context, email, password string) (string, string, error)
	VerifyOTP(ctx context.Context, email, otp string) error
	GetProfile(ctx context.Context, userID string) (*domain.UserProfile, error)
	UpdateProfile(ctx context.Context, profile *domain.UserProfile) error
	DeleteUser(ctx context.Context, userID string) error
	FilterDonors(ctx context.Context, filter domain.DonorFilter) ([]domain.DonorResponse, error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, otp, newPassword string) error
	UpdateDonorStatus(ctx context.Context, donorID, status string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	GetUsersByRole(ctx context.Context, role string) ([]domain.UserResponse, error)
	GetAllProfiles(ctx context.Context) ([]domain.UserProfile, error)
}
