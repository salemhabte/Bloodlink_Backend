package router

import (
	"bloodlink/Delivery/controller"
	domain "bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"
	"bloodlink/Infrastructure"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	userCtrl *controller.UserController,
	auth domainInterface.IAuthentication,
	campaignController *controller.CampaignController,
	donationController *controller.DonationController,
) *gin.Engine {

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:   []string{"Content-Length"},
	}))

	// Public Routes
	api := r.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", userCtrl.RegisterUser)
			authRoutes.POST("/login", userCtrl.HandleLogin)
			authRoutes.POST("/verify-otp", userCtrl.VerifyOTP)
			authRoutes.POST("/forgot-password", userCtrl.ForgotPassword)
			authRoutes.POST("/reset-password", userCtrl.ResetPassword)
			authRoutes.POST("/refresh-token", userCtrl.RefreshTokenHandler)
		}

		api.POST("/logout", Infrastructure.AuthMiddleware(auth), userCtrl.Logout)

		// Example Protected Routes (for verification)
		protectedRoutes := api.Group("/protected")
		protectedRoutes.Use(Infrastructure.AuthMiddleware(auth, domain.RoleDonor, domain.RoleBloodBankAdmin, domain.RoleBloodCollector, domain.RoleLabTech, domain.RoleHospitalAdmin))
		{
			protectedRoutes.GET("/profile", userCtrl.GetProfile)
			protectedRoutes.GET("/profile/:id", userCtrl.GetProfileByID)
			protectedRoutes.PATCH("/profile", userCtrl.UpdateProfile)
			protectedRoutes.DELETE("/user", userCtrl.DeleteUser)
			protectedRoutes.GET("/donors/filter", userCtrl.GetDonors)
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
	admin.Use(Infrastructure.AuthMiddleware(auth, domain.RoleBloodBankAdmin))
	{
		adminCampaigns := admin.Group("/campaigns")
		{
			adminCampaigns.POST("/", campaignController.CreateCampaign)
			adminCampaigns.PUT("/:id", campaignController.UpdateCampaign)
			adminCampaigns.DELETE("/:id", campaignController.DeleteCampaign)
		}

		adminDonors := admin.Group("/donors")
		{
			adminDonors.PUT("/:donor_id/status", userCtrl.UpdateDonorStatus)
		}

		adminUsers := admin.Group("/users")
		{
			adminUsers.GET("/filter", userCtrl.GetUsersByRole)
		}

		adminProfiles := admin.Group("/profiles")
		{
			adminProfiles.GET("/", userCtrl.GetAllProfiles)
		}
	}
	// Blood Collector Routes
bloodCollector := r.Group("/api/bloodcollector")
{
	bloodCollector.GET("/donor", donationController.SearchDonor)
	bloodCollector.POST("/donation", donationController.CreateDonation)
	bloodCollector.PUT("/donation/:id/status", donationController.UpdateDonationStatus)
}

	return r
}
