package Domain

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	RoleDonor          = "donor"
	RoleBloodBankAdmin = "bloodbankadmin"
	RoleHospitalAdmin  = "hospitaladmin"
	RoleBloodCollector = "bloodcollector"
	RoleLabTech        = "labtech"
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
	ID           string    `json:"id" db:"user_id"`
	FullName     string    `json:"full_name" db:"full_name"`
	Email        string    `json:"email" db:"email"`
	Phone        string    `json:"phone" db:"phone"`
	Password     string    `json:"password" db:"password"`
	Role         string    `json:"role" db:"role"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	OTP          string    `json:"otp" db:"otp"`
	RefreshToken string    `json:"-" db:"refresh_token"`
}

// UserProfile stores common display information for all roles
type UserProfile struct {
	ProfileID         string `json:"profile_id" db:"profile_id"`
	UserID            string `json:"user_id" db:"user_id"`
	FullName          string `json:"full_name" db:"full_name"`
	Phone             string `json:"phone" db:"phone"`
	Address           string `json:"address" db:"address"`
	ProfilePictureURL string `json:"profile_picture_url" db:"profile_picture_url"`
}

type Donor struct {
	DonorID          string `json:"donor_id" db:"donor_id"`
	UserID           string `json:"user_id" db:"user_id"`
	BloodType        string `json:"blood_type" db:"blood_type"`
	Status           string `json:"status" db:"status"`
	LastDonationDate string `json:"last_donation_date" db:"last_donation_date"`
}

type DonorResponse struct {
	DonorID   string `json:"donor_id" db:"donor_id"`
	UserID    string `json:"user_id" db:"user_id"`
	FullName  string `json:"full_name" db:"full_name"`
	Email     string `json:"email" db:"email"`
	Phone     string `json:"phone" db:"phone"`
	Address   string `json:"address" db:"address"`
	BloodType string `json:"blood_type" db:"blood_type"`
	Status    string `json:"status" db:"status"`
}

type DonorFilter struct {
	BloodType string `json:"blood_type"`
	Status    string `json:"status"`
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

// RegisterRequest represents the payload for user registration
type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

// LoginRequest represents the payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required"`
	OTP   string `json:"otp" binding:"required"`
}

type RefreshTokenRequestDTO struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UserResponse struct {
	ID        string    `json:"id" db:"user_id"`
	FullName  string    `json:"full_name" db:"full_name"`
	Email     string    `json:"email" db:"email"`
	Phone     string    `json:"phone" db:"phone"`
	Role      string    `json:"role" db:"role"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
