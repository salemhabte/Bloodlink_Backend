package Domain

import (
	"context"

	domain "bloodlink/Domain"
)

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
