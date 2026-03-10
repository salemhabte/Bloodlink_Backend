package router

import (
	"bloodlink/Delivery/controller"
	domain "bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"
	"bloodlink/Infrastructure"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	userCtrl *controller.UserController,
	auth domainInterface.IAuthentication,
	campaignController *controller.CampaignController,
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

	campaigns := r.Group("/api/campaigns")
	{
		campaigns.GET("/", campaignController.GetAllCampaigns)
		campaigns.GET("/:id", campaignController.GetCampaignByID)
		campaigns.GET("/search", campaignController.GetCampaignsByLocation)
	}
	//Campaign Routes Accessible by blood bank Admin

	admin := r.Group("/api/bloodbankadmin")
	{
		adminCampaigns := admin.Group("/campaigns")
		{
			adminCampaigns.POST("/", campaignController.CreateCampaign)
			adminCampaigns.PUT("/:id", campaignController.UpdateCampaign)
			adminCampaigns.DELETE("/:id", campaignController.DeleteCampaign)
		}
	}

	return r
}
