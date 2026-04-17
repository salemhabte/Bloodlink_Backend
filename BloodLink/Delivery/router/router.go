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
	labController *controller.LabController,
	inventoryController *controller.BloodInventoryController,
	hospitalController *controller.HospitalController,
	bloodReqController *controller.BloodRequestController,
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

		publicHospitals := api.Group("/public/hospitals")
		{
			publicHospitals.POST("/request-registration", hospitalController.SubmitRegistrationRequest)
		}

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

		adminHospitalRequests := admin.Group("/hospital-requests")
		{
			adminHospitalRequests.GET("/", hospitalController.GetPendingRequests)
			adminHospitalRequests.POST("/:id/approve", hospitalController.ApproveRequest)
			adminHospitalRequests.POST("/:id/reject", hospitalController.RejectRequest)
		}

		adminContracts := admin.Group("/contracts")
		{
			adminContracts.POST("/:id/sign", hospitalController.AdminSignContract)
			adminContracts.POST("/:id/reject", hospitalController.RejectContract)
		}

		adminTemplates := admin.Group("/contract-templates")
		{
			adminTemplates.GET("/", hospitalController.GetContractTemplates)
			adminTemplates.POST("/", hospitalController.CreateContractTemplate)
			adminTemplates.PUT("/:id", hospitalController.UpdateContractTemplate)
			adminTemplates.DELETE("/:id", hospitalController.DeleteContractTemplate)
		}

		adminProfiles := admin.Group("/profiles")
		{
			adminProfiles.GET("/", userCtrl.GetAllProfiles)
		}

		adminBloodRequests := admin.Group("/blood-requests")
		{
			adminBloodRequests.GET("/", bloodReqController.GetAllRequests)
			adminBloodRequests.PUT("/:id/status", bloodReqController.UpdateStatus)
		}
	}
	// Blood Collector Routes
bloodCollector := r.Group("/api/bloodcollector")
// bloodCollector.Use(Infrastructure.AuthMiddleware(auth, domain.RoleBloodCollector))
{
	bloodCollector.GET("/donors", donationController.GetPendingDonors)
bloodCollector.GET("/donor/:id", donationController.GetDonorByID)
bloodCollector.GET("/donor/search/pending", donationController.SearchPendingDonor)
    bloodCollector.GET("/donor/search", donationController.SearchDonor)
    bloodCollector.POST("/donation", donationController.CreateDonation)
    bloodCollector.GET("/donation", donationController.GetAllDonations)
    bloodCollector.GET("/donation/:id", donationController.GetDonationByID)
    bloodCollector.PUT("/donation/:id", donationController.UpdateDonation)
    bloodCollector.PUT("/donation/:id/status", donationController.UpdateDonationStatus)
}

lab := r.Group("/api/lab")
lab.Use(Infrastructure.AuthMiddleware(auth, domain.RoleLabTech))
{
	lab.POST("/tests", labController.SubmitTestResult)
	lab.GET("/tests/:donation_id", labController.GetTestResult)

	lab.GET("/donations/:donation_id", labController.GetDonation)

	lab.GET("/pending-tests", labController.GetPendingTests)
	lab.GET("/tests/history", labController.GetHistory)
	lab.GET("/tests", labController.FilterTests)

	lab.PUT("/tests/:donation_id", labController.UpdateTest)
	lab.POST("/tests/:donation_id/reject", labController.RejectBlood)
}
adminInventory := r.Group("/api/inventory")
adminInventory.Use(Infrastructure.AuthMiddleware(auth, domain.RoleBloodBankAdmin))
{
	adminInventory.GET("/", inventoryController.GetAll)
	adminInventory.GET("/stats", inventoryController.GetStats)
	adminInventory.GET("/filter", inventoryController.Filter)
	adminInventory.GET("/export/csv", inventoryController.ExportCSV)
	adminInventory.GET("/export/pdf", inventoryController.ExportPDF)
	adminInventory.GET("/:id/details", inventoryController.GetFullDetails)

	adminInventory.PUT("/:id/status", inventoryController.UpdateStatus)
	adminInventory.DELETE("/:id", inventoryController.Delete)
}
labInventory := r.Group("/api/lab/inventory")
labInventory.Use(Infrastructure.AuthMiddleware(auth, domain.RoleLabTech))
{
	labInventory.GET("/", inventoryController.GetAll)
	labInventory.GET("/filter", inventoryController.Filter)
	labInventory.GET("/:id/details", inventoryController.GetFullDetails)
}

hospitalGrp := r.Group("/api/hospitaladmin")
hospitalGrp.Use(Infrastructure.AuthMiddleware(auth, domain.RoleHospitalAdmin))
{
	hContracts := hospitalGrp.Group("/contracts")
	{
		hContracts.POST("/:id/sign", hospitalController.HospitalSignContract)
		hContracts.POST("/:id/reject", hospitalController.RejectContract)
	}

	hBloodReqs := hospitalGrp.Group("/blood-requests")
	{
		hBloodReqs.POST("/", bloodReqController.CreateBloodRequest)
		hBloodReqs.GET("/", bloodReqController.GetHospitalRequests)
	}
}

	return r
}
