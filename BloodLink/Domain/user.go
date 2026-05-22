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
	ID           string     `json:"id" db:"user_id"`
	FullName     string     `json:"full_name" db:"full_name"`
	Email        string     `json:"email" db:"email"`
	Phone        string     `json:"phone" db:"phone"`
	Password     string     `json:"password" db:"password"`
	Role         string     `json:"role" db:"role"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	OTP          string     `json:"otp" db:"otp"`
	OTPExpiresAt *time.Time `json:"otp_expires_at,omitempty" db:"otp_expires_at"`
	RefreshToken string     `json:"-" db:"refresh_token"`
}

type UserProfile struct {
	ProfileID         string   `json:"profile_id" db:"profile_id"`
	UserID            string   `json:"user_id" db:"user_id"`
	FullName          string   `json:"full_name" db:"full_name"`
	Phone             string   `json:"phone" db:"phone"`
	Address           string   `json:"address" db:"address"`
	ProfilePictureURL string   `json:"profile_picture_url" db:"profile_picture_url"`
	Latitude          *float64 `json:"latitude" db:"latitude"`
	Longitude         *float64 `json:"longitude" db:"longitude"`
}

type DonorEligibility struct {
	IsEligible         bool   `json:"is_eligible"`
	EligibilityStatus  string `json:"eligibility_status"` // "Eligible" or "Not Eligible"
	EligibilityMessage string `json:"eligibility_message"`
	CountdownDays      int    `json:"countdown_days,omitempty"`
}

type ProfileResponse struct {
	UserProfile
	DonorInfo   *Donor            `json:"donor_info,omitempty"`
	Eligibility *DonorEligibility `json:"eligibility,omitempty"`
}

type Donor struct {
	DonorID          string    `json:"donor_id" db:"donor_id"`
	UserID           string    `json:"user_id" db:"user_id"`
	BloodType        string    `json:"blood_type" db:"blood_type"`
	OverallStatus    string    `json:"overall_status" db:"overall_status"`
	DateOfBirth      time.Time `json:"date_of_birth" db:"date_of_birth"`
	LastDonationDate string    `json:"last_donation_date" db:"last_donation_date"`
}

type DonorResponse struct {
	DonorID          string     `json:"donor_id" db:"donor_id"`
	UserID           string     `json:"user_id" db:"user_id"`
	FullName         string     `json:"full_name" db:"full_name"`
	Email            string     `json:"email" db:"email"`
	Phone            string     `json:"phone" db:"phone"`
	Address          string     `json:"address" db:"address"`
	BloodType        *string    `json:"blood_type" db:"blood_type"`
	Status           string     `json:"status" db:"status"`
	OverallStatus    string     `json:"overall_status" db:"overall_status"`
	RegistrationDate time.Time  `json:"registration_date" db:"created_at"`
	LastDonationDate *time.Time `json:"last_donation_date,omitempty" db:"last_donation"`
}

type EligibleDonorsResponse struct {
	TotalEligible     int             `json:"total_eligible"`
	ReturningEligible int             `json:"returning_eligible"`
	NewEligibleDonors int             `json:"new_eligible_donors"`
	Donors            []DonorResponse `json:"donors"`
}

type AllDonorsResponse struct {
	TotalDonors         int             `json:"total_donors"`
	Cleared             int             `json:"cleared"`
	TemporarilyDeferred int             `json:"temporarily_deferred"`
	PermanentlyDeferred int             `json:"permanently_deferred"`
	Donors              []DonorResponse `json:"donors"`
}

type DonorFilter struct {
	BloodType     string `json:"blood_type"`
	OverallStatus string `json:"overall_status"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Status        string `json:"status"` // is_active
	IsEligible    *bool  `json:"is_eligible"`
	IsNewDonor    *bool  `json:"is_new_donor"`
}

type UserFilter struct {
	Role      string `json:"role"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type ProfileFilter struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
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

type RegisterDonorRequest struct {
	FullName  string    `json:"full_name" binding:"required"`
	Email     string    `json:"email" binding:"required"`
	Phone     string    `json:"phone" binding:"required"`
	Password  string    `json:"password" binding:"required,min=8"`
	Address   string    `json:"address" binding:"required"`
	BirthDate time.Time `json:"birth_date" binding:"required"`
	Latitude  float64   `json:"latitude,string"`
	Longitude float64   `json:"longitude,string"`
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

type SendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
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

type SendGridRequest struct {
	Personalizations []Personalization `json:"personalizations"`
	From             EmailAddress      `json:"from"`
	Subject          string            `json:"subject"`
	Content          []Content         `json:"content"`
}

type Personalization struct {
	To      []EmailAddress `json:"to"`
	Subject string         `json:"subject,omitempty"`
}

type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type Content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
