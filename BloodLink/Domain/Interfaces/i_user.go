package Domain

import (
	"context"

	domain "bloodlink/Domain"
)

// IIndividualRepository now uses context.Context and works with the domain model.
type IIndividualRepository interface {
	CreateIndividual(ctx context.Context, individual *domain.UserProfile) (*domain.UserProfile, error)
	FindByEmail(ctx context.Context, email string) (*domain.UserProfile, error)
	FindByID(ctx context.Context, individualID string) (*domain.UserProfile, error)
	FindUser(ctx context.Context, userID string) (*domain.UserProfile, error)
	UpdateIndividual(ctx context.Context, individualID string, updates map[string]interface{}) error // for the time being
	UpdateResetOTP(ctx context.Context, email, otp string) error
	VerifyResetOTP(ctx context.Context, email, otp string) error
	UpdatePasswordByEmail(ctx context.Context, email, newHashedPassword string) error
	DeleteIndividual(ctx context.Context, individualID string) error
	DeleteRefreshToken(ctx context.Context, userID string) error
	UpdateProfile(ctx context.Context, email string, updateData map[string]interface{}) error
}

// we should use GetByID instead of GetByEmail for performance
type IOTPRepository interface {
	CreateUnverifiedUser(ctx context.Context, unverifiedUser *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	DeleteByID(ctx context.Context, userID string) error
}

type IUserValidation interface {
	IsValidEmail(email string) bool
	IsStrongPassword(password string) bool
	Hashpassword(password string) string
	ComparePassword(userPassword, password string) error
}
