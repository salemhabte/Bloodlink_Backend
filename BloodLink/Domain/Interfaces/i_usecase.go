package Domain

import (
	"context"
	domain "bloodlink/Domain"
)

type IUserUseCase interface {
	ReSendAccessToken(jwtToken string) (string, error) // (accessTokenString, error)
	ValidOTPRequest(emailOtp *domain.EmailOTP) (*domain.User, error)
	StoreUserInOTPColl(user *domain.User) (error)
	StoreUserInMainColl(user *domain.User) (*domain.User, error)
	Login(email, password string) (string,string, error)
	SendResetOTP(ctx context.Context, email string) error
	Logout(ctx context.Context, user string) error
	ResetPassword(ctx context.Context, email, otp, newPassword string) error
	GetProfile(ctx context.Context, userID string) (*domain.UserProfile, error)
	UpdateProfile(ctx context.Context, userID string, updateReq *domain.UserProfile) error
}