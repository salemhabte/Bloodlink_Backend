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
	campaignAnalyticsController *controller.CampaignAnalyticsController,
	collectorAnalyticsController *controller.CollectorAnalyticsController,
	labAnalyticsController *controller.LabAnalyticsController,
	adminAnalyticsController *controller.AdminAnalyticsController,
	badgeController *controller.DonorBadgeController,
	emergencyController *controller.EmergencyRequestController,
	donorBloodReqController *controller.DonorBloodRequestController,
	notifController *controller.NotificationController,
	auditLogController *controller.AuditLogController,
) *gin.Engine {

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "https://blood-link-frontend-qw7r.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Public Routes
	api := r.Group("/api")
	{
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/register", userCtrl.RegisterUser)
			authRoutes.POST("/register-donor", userCtrl.RegisterDonor)
			authRoutes.POST("/login", userCtrl.HandleLogin)
			authRoutes.POST("/send-otp", userCtrl.SendOTP)
			authRoutes.POST("/resend-otp", userCtrl.ResendOTP)
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

		api.GET("/emergencies/published", emergencyController.GetPublishedEmergencies)
	}

	notifications := r.Group("/api/notifications")
	notifications.Use(Infrastructure.AuthMiddleware(auth)) // All logged in users can see their own notifications
	{
		notifications.GET("/", notifController.GetMyNotifications)
		notifications.PUT("/:id/read", notifController.MarkAsRead)
		notifications.PUT("/read-all", notifController.MarkAllAsRead)
	}

	campaigns := r.Group("/api/campaigns")
	{
		campaigns.GET("/", campaignController.GetAllCampaigns)
		campaigns.GET("/:id", campaignController.GetCampaignByID)
	}
	//Campaign Routes Accessible by blood bank Admin

	admin := r.Group("/api/bloodbankadmin")
	admin.Use(Infrastructure.AuthMiddleware(auth, domain.RoleBloodBankAdmin))
	{
		admin.GET("/leaderboard", badgeController.GetLeaderboard)
		admin.GET("/badges", badgeController.GetAllBadges)
		admin.GET("/donor-blood-requests", donorBloodReqController.GetAllAdminRequests) // filtered + sorted by donations
		admin.POST("/donor-blood-requests/expire-stale", donorBloodReqController.ExpireStaleRequests)
		admin.PUT("/blood-requests/:id/approve", donorBloodReqController.ApproveRequest)
		admin.PUT("/blood-requests/:id/fulfill", donorBloodReqController.FulfillRequest)
		admin.PUT("/blood-requests/:id/reject", donorBloodReqController.RejectRequest)
		adminCampaigns := admin.Group("/campaigns")
		{
			adminCampaigns.POST("", campaignController.CreateCampaign)
			adminCampaigns.POST("/", campaignController.CreateCampaign)
			adminCampaigns.PUT("/:id", campaignController.UpdateCampaign)
			adminCampaigns.DELETE("/:id", campaignController.DeleteCampaign)
		}

		adminDonors := admin.Group("/donors")
		{
			adminDonors.GET("/", userCtrl.GetDonors)
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
			adminContracts.GET("/signed", hospitalController.GetSignedContracts)
			adminContracts.GET("/:id", hospitalController.GetContractByID)
			adminContracts.GET("/:id/download", hospitalController.DownloadContract)
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
			adminBloodRequests.GET("", bloodReqController.GetAllRequests)
			adminBloodRequests.GET("/", bloodReqController.GetAllRequests)
			adminBloodRequests.POST("/:id/approve", bloodReqController.ApproveRequest)
			adminBloodRequests.POST("/:id/reject", bloodReqController.RejectRequest)
		}

		// Mark all reserved units as USED when hospital collects in person
		admin.PATCH("/reserved/:id/mark-used", bloodReqController.MarkRequestUnitsAsUsed)

		adminEmergencies := admin.Group("/emergencies")
		{
			adminEmergencies.GET("/", emergencyController.GetAllEmergencies)
			adminEmergencies.POST("/manual", emergencyController.CreateManualEmergency)
			adminEmergencies.POST("/:id/publish", emergencyController.PublishEmergency)
			adminEmergencies.POST("/:id/reject", emergencyController.RejectEmergency)
		}

		admin.GET("/hospitals", hospitalController.GetAllHospitals)
		admin.GET("/audit-logs", auditLogController.GetLogs)
		admin.GET("/audit-logs/export", auditLogController.ExportLogs)
		admin.GET("/audit-logs/:id", auditLogController.GetLogByID)
		admin.DELETE("/audit-logs/:id", auditLogController.DeleteLog)
	}

	adminCampaignsRead := r.Group("/api/bloodbankadmin/campaigns")
	adminCampaignsRead.Use(Infrastructure.AuthMiddleware(auth, domain.RoleBloodBankAdmin, domain.RoleBloodCollector))
	{
		adminCampaignsRead.GET("", campaignController.GetAllAdminCampaigns)
		adminCampaignsRead.GET("/", campaignController.GetAllAdminCampaigns)
	}

	// Donor Routes
	donor := r.Group("/api/donor")
	{
		donor.Use(Infrastructure.AuthMiddleware(auth, domain.RoleDonor))
		{
			donor.GET("/donations", donationController.GetAllDonationsByDonor)
			donor.GET("/test-result/latest", labController.GetLatestTestResultByDonor)
			donor.GET("/badges", badgeController.GetMyBadges)
			donor.GET("/leaderboard", badgeController.GetLeaderboard)
			donor.GET("/emergencies", emergencyController.GetEmergenciesForDonor)
			donor.POST("/blood-request", donorBloodReqController.CreateRequest)
			donor.GET("/my-requests", donorBloodReqController.GetMyRequests)
			donor.POST("/location", userCtrl.UpdateLocation)

		}
	}

	// Blood Collector Routes
	bloodCollector := r.Group("/api/bloodcollector")
	bloodCollector.Use(Infrastructure.AuthMiddleware(auth, domain.RoleBloodCollector))
	{
		bloodCollector.GET("/eligible-donors", userCtrl.GetEligibleDonors)
		bloodCollector.GET("/all-donors", userCtrl.GetDonors)
		bloodCollector.GET("/eligible-donor/:id", donationController.GetDonorByID)

		bloodCollector.POST("/donation", donationController.CreateDonation)
		bloodCollector.GET("/donation", donationController.GetAllDonations)
		bloodCollector.GET("/donation/:id", donationController.GetDonationByID)
		bloodCollector.PUT("/donation/:id", donationController.UpdateDonation)
		bloodCollector.GET("/donation/my", donationController.GetMyDonations)
	}

	lab := r.Group("/api/lab")
	lab.Use(Infrastructure.AuthMiddleware(auth, domain.RoleLabTech))
	{
		lab.POST("/tests", labController.SubmitTestResult)
		lab.PUT("/tests/:donation_id", labController.UpdateTest)
		lab.PATCH("/tests/:donation_id/reject", labController.RejectBlood)

		// SINGLE test
		lab.GET("/tests/:donation_id", labController.GetTestResult)

		// PRIMARY TEST LIST (History + Filters)
		lab.GET("/all-tests", labController.GetAllTestHistory)

		// MY tests
		lab.GET("/tests/my", labController.GetMyTests)

		//  donations (fetch before testing)
		lab.GET("/pending-donations", labController.GetPendingDonations)
		lab.GET("/pending-donations/:donation_id", labController.GetPendingDonation)
	}

	inventory := r.Group("/api/inventory")
	inventory.Use(Infrastructure.AuthMiddleware(auth, domain.RoleBloodBankAdmin, domain.RoleLabTech))
	{
		// Shared Routes
		inventory.GET("/", inventoryController.GetAll)
		inventory.GET("/:id", inventoryController.GetByID)
		inventory.GET("/stats", inventoryController.GetStats)
		inventory.GET("/export/csv", inventoryController.ExportCSV)
		inventory.GET("/export/pdf", inventoryController.ExportPDF)
		inventory.GET("/:id/details", inventoryController.GetFullDetails)
		inventory.DELETE("/:id", inventoryController.Delete)

		// Admin-Only Routes
		adminActions := inventory.Group("/")
		adminActions.Use(Infrastructure.AuthMiddleware(auth, domain.RoleBloodBankAdmin))
		{
			adminActions.GET("/reserved/:hospital_id", inventoryController.GetReservedByHospital)
			adminActions.GET("/reserved-request/:request_id", inventoryController.GetReservedByRequest)
			adminActions.PUT("/:id/used", inventoryController.MarkUsed)
		}

		// Lab-Only Routes
		labActions := inventory.Group("/")
		labActions.Use(Infrastructure.AuthMiddleware(auth, domain.RoleLabTech))
		{
			labActions.POST("/:id/convert-cryo", inventoryController.ConvertPlasma)
		}
	}
	analytics := r.Group("/api/analytics/campaigns")

	analytics.GET("/dashboard", campaignAnalyticsController.GetDashboard)
	analytics.GET("/:id", campaignAnalyticsController.GetCampaignReport)
	analytics.GET("/", campaignAnalyticsController.GetAllReports)

	collectorAnalytics := r.Group("/api/analytics/collector")
	collectorAnalytics.Use(Infrastructure.AuthMiddleware(auth, domain.RoleBloodCollector))
	{
		collectorAnalytics.GET("/kpi", collectorAnalyticsController.GetKPI)
		collectorAnalytics.GET("/today", collectorAnalyticsController.GetTodayStats)
		collectorAnalytics.GET("/donor-insights", collectorAnalyticsController.GetDonorInsights)
	}

	labAnalytics := r.Group("/api/analytics/lab")
	labAnalytics.Use(Infrastructure.AuthMiddleware(auth, domain.RoleLabTech))
	{
		labAnalytics.GET("/dashboard", labAnalyticsController.GetDashboard)
	}
	adminAnalytics := r.Group("/api/analytics/admin")
	adminAnalytics.Use(Infrastructure.AuthMiddleware(auth, domain.RoleBloodBankAdmin))
	{
		// FULL DASHBOARD
		adminAnalytics.GET("/dashboard", adminAnalyticsController.GetDashboard)

		// OPTIONAL SPLIT ENDPOINTS
		adminAnalytics.GET("/donors", adminAnalyticsController.GetDonorSummary)
		adminAnalytics.GET("/screening", adminAnalyticsController.GetScreeningSummary)
		adminAnalytics.GET("/collectors", adminAnalyticsController.GetCollectorSummary)
		adminAnalytics.GET("/lab", adminAnalyticsController.GetLabSummary)
		adminAnalytics.GET("/inventory", adminAnalyticsController.GetInventorySummary)
	}

	hospitalGrp := r.Group("/api/hospitaladmin")
	hospitalGrp.Use(Infrastructure.AuthMiddleware(auth, domain.RoleHospitalAdmin))
	{
		hospitalGrp.GET("/dashboard", hospitalController.GetHospitalDashboard)

		hContracts := hospitalGrp.Group("/contracts")
		{
			hContracts.GET("/", hospitalController.GetMyContracts)
			hContracts.GET("/latest", hospitalController.GetMyLatestContract)
			hContracts.GET("/:id", hospitalController.GetContractByID)
			hContracts.GET("/:id/download", hospitalController.DownloadContract)
			hContracts.POST("/:id/sign", hospitalController.HospitalSignContract)
			hContracts.POST("/:id/reject", hospitalController.RejectContract)
		}

		hBloodReqs := hospitalGrp.Group("/blood-requests")
		{
			hBloodReqs.POST("/", bloodReqController.CreateBloodRequest)
			hBloodReqs.GET("/", bloodReqController.GetHospitalRequests)
		}

		hDonations := hospitalGrp.Group("/donations")
		{
			hDonations.POST("/confirm", hospitalController.ConfirmDonation)
			hDonations.GET("/donor-profile", hospitalController.GetDonorProfileByPhone)
		}
	}

	r.Static("/uploads", "./uploads")

	return r
}
