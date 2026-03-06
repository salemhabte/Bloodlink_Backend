package Domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	Donor          = "donor"
	BloodBankAdmin = "bloodbankadmin"
	HospitalAdmin  = "hospitaladmin"
	BloodCollector = "bloodcollector"
	LabTech        = "labtech"
)

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

type UserClaims struct {
	UserID      string `json:"id"`
	Email       string `json:"email"`
	IsVerified  bool   `json:"is_verified"`
	AccountType string `json:"account_type"`
	TokenType   string `json:"token_type"` // The requested field to identify the token's type
	jwt.RegisteredClaims
}

type User struct {
	ID        string    `json:"id" db:"user_id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password" db:"password"`
	Role      string    `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	OTP       string    `json:"otp" db:"otp"`
}

type UserProfile struct {
	ProfileID         string `json:"profile_id" db:"profile_id"`
	UserID            string `json:"user_id" db:"user_id"`
	Email             string `json:"email" db:"email"`
	Password          string `json:"password" db:"password"`
	FullName          string `json:"full_name" db:"full_name"`
	Phone             string `json:"phone" db:"phone"`
	Address           string `json:"address" db:"address"`
	ProfilePictureURL string `json:"profile_picture_url" db:"profile_picture_url"`
}

type EmailOTP struct {
	Email string `json:"email" bson:"email"`
	OTP   string `json:"otp" bson:"otp"`
}

type ForgotPasswordRequestDTO struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequestDTO struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=50"`
}
