package Domain

import (
	"context"
	domain "bloodlink/Domain"
)

type IUserUseCase interface {
	RegisterUser(ctx context.Context, user *domain.User) error
	Login(ctx context.Context, email, password string) (string, string, error)
	VerifyOTP(ctx context.Context, email, otp string) error
	GetProfile(ctx context.Context, userID string) (*domain.UserProfile, error)
	UpdateProfile(ctx context.Context, profile *domain.UserProfile) error
	DeleteUser(ctx context.Context, userID string) error
}