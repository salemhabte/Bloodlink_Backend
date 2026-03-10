package router

import (
	"bloodlink/Delivery/controller"
	"bloodlink/Infrastructure"
	domain "bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	userCtrl *controller.UserController,
	auth domainInterface.IAuthentication,
) *gin.Engine {

	r := gin.Default()

	// Public Routes
	api := r.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", userCtrl.RegisterUser)
			authRoutes.POST("/login", userCtrl.HandleLogin)
			authRoutes.POST("/verify-otp", userCtrl.VerifyOTP)
		}

		// Example Protected Routes (for verification)
		protectedRoutes := api.Group("/protected")
		protectedRoutes.Use(Infrastructure.AuthMiddleware(auth, domain.RoleDonor, domain.RoleBloodBankAdmin, domain.RoleBloodCollector, domain.RoleLabTech, domain.RoleHospitalAdmin))
		{
			protectedRoutes.GET("/profile", userCtrl.GetProfile)
			protectedRoutes.PUT("/profile", userCtrl.UpdateProfile)
			protectedRoutes.DELETE("/user", userCtrl.DeleteUser)
		}
	}

	return r
}
