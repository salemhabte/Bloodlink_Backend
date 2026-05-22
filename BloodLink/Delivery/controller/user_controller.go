package controller

import (
	"context"
	"log"
	"net/http"
	"strings"

	domain "bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	UserUseCase domainInterface.IUserUseCase
}

func NewUserController(userUseCase domainInterface.IUserUseCase) *UserController {
	return &UserController{
		UserUseCase: userUseCase,
	}
}

func (c *UserController) RegisterUser(ctx *gin.Context) {
	var req domain.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &domain.User{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		FullName: req.FullName,
		Phone:    req.Phone,
	}

	// For standard API requests, contextual timeouts are good practice
	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.RegisterUser(cCtx, user); err != nil {
		log.Printf("[ERROR] RegisterUser failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

func (c *UserController) RegisterDonor(ctx *gin.Context) {
	var req domain.RegisterDonorRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.RegisterDonor(cCtx, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Donor registered successfully. You can now login.",
	})
}

func (c *UserController) HandleLogin(ctx *gin.Context) {
	var req domain.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	accessToken, refreshToken, role, userID, otpNeeded, err := c.UserUseCase.Login(cCtx, req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"role":          role,
		"user_id":       userID,
	}

	// Add otp_needed for specific roles
	if role == domain.RoleBloodCollector || role == domain.RoleLabTech {
		response["otp_needed"] = otpNeeded
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *UserController) VerifyOTP(ctx *gin.Context) {
	var req domain.VerifyOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.VerifyOTP(cCtx, req.Email, req.OTP); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully. Your profile has been created.",
	})
}

func (c *UserController) SendOTP(ctx *gin.Context) {
	var req domain.SendOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.SendOTP(cCtx, req.Email); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OTP sent to your email"})
}

func (c *UserController) ResendOTP(ctx *gin.Context) {
	var req domain.SendOTPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.ResendOTP(cCtx, req.Email); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "OTP resent to your email"})
}

func (c *UserController) GetProfile(ctx *gin.Context) {
	userID := ctx.GetString("userID") // Match middleware key
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in context"})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	profile, err := c.UserUseCase.GetProfile(cCtx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if profile == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	ctx.JSON(http.StatusOK, profile)
}

// GetProfileByID fetches a user's profile based on the ID passed in the path
func (c *UserController) GetProfileByID(ctx *gin.Context) {
	userID := ctx.Param("id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	profile, err := c.UserUseCase.GetProfile(cCtx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if profile == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}

	ctx.JSON(http.StatusOK, profile)
}

// GetAllProfiles fetches all user profiles in the system
func (c *UserController) GetAllProfiles(ctx *gin.Context) {
	filter := domain.ProfileFilter{
		StartDate: ctx.Query("start_date"),
		EndDate:   ctx.Query("end_date"),
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	profiles, err := c.UserUseCase.GetAllProfiles(cCtx, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, profiles)
}

func (c *UserController) UpdateProfile(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in context"})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	// 1. Fetch the existing profile
	existingProfile, err := c.UserUseCase.GetProfile(cCtx, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing profile"})
		return
	}
	if existingProfile == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Profile not found to update"})
		return
	}

	// 2. Read the partial updates from the request
	var updates map[string]interface{}
	if err := ctx.ShouldBindJSON(&updates); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Apply only the provided updates
	if val, ok := updates["full_name"].(string); ok {
		existingProfile.FullName = val
	}
	if val, ok := updates["phone"].(string); ok {
		existingProfile.Phone = val
	}
	if val, ok := updates["address"].(string); ok {
		existingProfile.Address = val
	}
	if val, ok := updates["profile_picture_url"].(string); ok {
		existingProfile.ProfilePictureURL = val
	}

	// 4. Save the merged profile back to the database
	if err := c.UserUseCase.UpdateProfile(cCtx, &existingProfile.UserProfile); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"profile": existingProfile,
	})
}

func (c *UserController) DeleteUser(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in context"})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.DeleteUser(cCtx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User and all related data deleted successfully"})
}

func (c *UserController) UpdateDonorStatus(ctx *gin.Context) {
	donorID := ctx.Param("donor_id")
	if donorID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "donor_id is required"})
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.UpdateDonorStatus(cCtx, donorID, body.Status); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Donor status updated to " + body.Status})
}

func (c *UserController) ForgotPassword(ctx *gin.Context) {
	var req domain.ForgotPasswordRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.ForgotPassword(cCtx, req.Email); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password reset OTP sent to your email"})
}

func (c *UserController) ResetPassword(ctx *gin.Context) {
	var req domain.ResetPasswordRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.ResetPassword(cCtx, req.Email, req.OTP, req.NewPassword); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (c *UserController) GetDonors(ctx *gin.Context) {
	bloodType := ctx.Query("blood_type")
	if bloodType != "" {
		bloodType = strings.ReplaceAll(bloodType, " ", "+")
	}

	filter := domain.DonorFilter{
		BloodType:     bloodType,
		OverallStatus: ctx.Query("overall_status"),
		StartDate:     ctx.Query("start_date"),
		EndDate:       ctx.Query("end_date"),
		Status:        ctx.Query("status"),
	}

	if isEligible := ctx.Query("is_eligible"); isEligible != "" {
		val := isEligible == "true"
		filter.IsEligible = &val
	}
	if isNewDonor := ctx.Query("is_new_donor"); isNewDonor != "" {
		val := isNewDonor == "true"
		filter.IsNewDonor = &val
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	donors, err := c.UserUseCase.FilterDonors(cCtx, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, donors)
}

func (c *UserController) RefreshTokenHandler(ctx *gin.Context) {
	var req domain.RefreshTokenRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	accessToken, refreshToken, err := c.UserUseCase.RefreshToken(cCtx, req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "tokens refreshed successfully",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// GetUsersByRole handles fetching users filtered by their role
func (c *UserController) GetUsersByRole(ctx *gin.Context) {
	filter := domain.UserFilter{
		Role:      ctx.Query("role"),
		StartDate: ctx.Query("start_date"),
		EndDate:   ctx.Query("end_date"),
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	users, err := c.UserUseCase.GetUsersByRole(cCtx, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (c *UserController) Logout(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in context"})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.Logout(cCtx, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (c *UserController) GetEligibleDonors(ctx *gin.Context) {
	query := ctx.Query("q")

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	donors, err := c.UserUseCase.GetEligibleDonors(cCtx, query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, donors)
}

func (c *UserController) UpdateLocation(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in context"})
		return
	}

	var req struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.UpdateLocation(cCtx, userID, req.Latitude, req.Longitude); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Location updated successfully"})
}
