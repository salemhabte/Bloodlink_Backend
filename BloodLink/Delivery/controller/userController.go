package controller

import (
	"fmt"
	"log"
	"net/http"

	domain "bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"
	infrastracture "bloodlink/Infrastructure"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userUseCase domainInterface.IUserUseCase
	

}

// RegisterIndividual implements domain.IUserController.
func (u *UserController) RegisterUser(ctx *gin.Context) {
	var unverifiedUser domain.User
	if err := ctx.ShouldBindJSON(&unverifiedUser); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}
	if !infrastracture.NewPasswordService().IsValidEmail(unverifiedUser.Email){
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data": gin.H{
				"error": "wrong email format",
			},
		})
		return
	}
	if !infrastracture.NewPasswordService().IsStrongPassword(unverifiedUser.Password){
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data": gin.H{
				"error": "your password is not strong Enough",
			},
		})
		return
	}
	
	unverifiedUser.Password= infrastracture.NewPasswordService().Hashpassword(unverifiedUser.Password)
	otp:= infrastracture.GenerateOTP()
	unverifiedUser.OTP=otp
	infrastracture.SendOTP(unverifiedUser.Email,otp)
	log.Print("=========",unverifiedUser)
	err := u.userUseCase.StoreUserInOTPColl(&unverifiedUser)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}
	// send the otp using email
	ctx.IndentedJSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Otp has been sent. Please verify your email.",
		},
	})
}

// VerfiyOTPRequest implements domain.IUserController.
func (u *UserController) VerfiyOTPRequest(ctx *gin.Context) {

	var emailOTP domain.EmailOTP
	if err := ctx.ShouldBindJSON(&emailOTP); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}

	userInfo, err := u.userUseCase.ValidOTPRequest(&emailOTP)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{ // Changed to StatusBadRequest for invalid requests
			"success": false,
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}

	_, err = u.userUseCase.StoreUserInMainColl(userInfo)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"data": gin.H{
				"error": err.Error(),
			},
		})
		return
	}
	
	ctx.IndentedJSON(http.StatusAccepted, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Email successfully verified.",
		},
	})
}

// RefreshTokenHandler implements domain.IController.
func (u *UserController) RefreshTokenHandler(ctx *gin.Context) {
	// get refresh token from cookie
	refreshToken, err := ctx.Cookie("WEKIL-API-REFRESH-TOKEN") //! don't forget to make the string in the cookie to a const
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh token cookie not found"})
		return
	}
	// validate refresh token and if the refresh token is valid then
	accessToken, err := u.userUseCase.ReSendAccessToken(refreshToken)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	// send the access token to the user and send accepted status
	ctx.Header("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	ctx.IndentedJSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Refreshed successfully. Tokens sent in header and cookie.",
		},
	})

}

func (uc *UserController) HandleLogin(ctx *gin.Context) {

	var user *domain.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Invalid request payload","success": false,
		})
		return
	}
	if user.Email == "" || user.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload","success": false,})
		return
	}
	accessToken,refreshToken, err := uc.userUseCase.Login(user.Email, user.Password)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	ctx.SetCookie(
		"WEKIL-API-REFRESH-TOKEN",
		refreshToken,
		60*60*24*7,      // 7 days in seconds
		"/",      // cookie path
		"",              // domain ("" means current domain)
		true,            // secure
		true,            // httpOnly
	)

	ctx.Header("Authorization", "Bearer "+accessToken)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "login successful",
		},
	})
}

func (uc *UserController) UpdateProfile(ctx *gin.Context) {
	var updateReq domain.UserProfile

	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input","success": false,})
		return
	}

	email := ctx.GetString("email")
	err := uc.userUseCase.UpdateProfile(ctx, email, &updateReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile","success": false,})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully","success": true,})
}

func (uc *UserController) GetProfile(ctx *gin.Context) {
	email := ctx.GetString("email")
	log.Println("id============---------:", email)
	profile, err := uc.userUseCase.GetProfile(ctx, email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve profile!!","success": false,})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

func (uc UserController) Logout(ctx *gin.Context) {
		userID := ctx.GetString("user_id")
		log.Println("id============:", userID)

		err := uc.userUseCase.Logout(ctx, userID)
			if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed*****","success": false,})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
		"data": gin.H{
			"message": "logged out successfully",
		},
	})
	}


func (u *UserController) SendResetOTP(c *gin.Context) {
	var req domain.ForgotPasswordRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),"success": false,})
		return
	}

	log.Println("Forgot password request received for:", req.Email)

	err := u.userUseCase.SendResetOTP(c, req.Email)
	if err != nil {
		log.Println("SendResetOTP error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset OTP","success": false,})
		return
	}

	c.JSON(http.StatusOK, gin.H{
			"success": true,
		"data": gin.H{
			"message": "Reset OTP sent to your email address",
		},
	})
}


func (uc *UserController) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input","success": false,})
		return
	}

	err := uc.userUseCase.ResetPassword(c, req.Email, req.OTP, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(),"success": false,})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Password reset successfully",
		},
		})
}