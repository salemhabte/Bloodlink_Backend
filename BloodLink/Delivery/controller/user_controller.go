package controller

import (
	"context"
	"log"
	"net/http"

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

func (c *UserController) HandleLogin(ctx *gin.Context) {
	var req domain.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	accessToken, refreshToken, err := c.UserUseCase.Login(cCtx, req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
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

func (c *UserController) UpdateProfile(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found in context"})
		return
	}

	var profile domain.UserProfile
	if err := ctx.ShouldBindJSON(&profile); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile.UserID = userID // Security: ensure they only update their own profile

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.UserUseCase.UpdateProfile(cCtx, &profile); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
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
