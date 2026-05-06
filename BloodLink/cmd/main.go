package main

import (
	"bloodlink/Delivery/controller"
	"bloodlink/Delivery/router"
	"bloodlink/Infrastructure"
	"bloodlink/Repository"
	"bloodlink/Usecase"
	"bloodlink/Jobs"

	"bloodlink/config"
	"log"
)

func main() {
	// 1. Initialize Configuration
	config.InitEnv()

	// 2. Connect to Database (MySQL)
	Repository.ConnectDB()
	db := Repository.DB

	// Run Database Migrations (Auto-Create Tables)
	Repository.RunMigrations()

	// 3. Initialize Infrastructure Services
	passwordService := Infrastructure.NewPasswordService()
	jwtService := Infrastructure.NewJWTAuthentication(config.JWTSECRET)

	// 4. Initialize Auth System
	userRepo := Repository.NewUserRepository(db)
	profileRepo := Repository.NewProfileRepository(db)
	userUseCase := Usecase.NewUserUseCase(userRepo, profileRepo, jwtService, passwordService)
	userController := controller.NewUserController(userUseCase)
	campaignRepo := Repository.NewCampaignRepository(db)
	donationRepo := Repository.NewDonationRepository(db)
	labRepo := Repository.NewLabRepository(db)
	inventoryRepo := Repository.NewBloodInventoryRepository(db)
	hospitalRepo := Repository.NewHospitalRepository(db)
	campaignAnalyticsRepo := Repository.NewCampaignAnalyticsRepository(db)
	collectorAnalyticsRepo := Repository.NewCollectorAnalyticsRepository(db)
	labAnalyticsRepo := Repository.NewLabAnalyticsRepository(db)
	adminAnalyticsRepo := Repository.NewAdminAnalyticsRepository(db)
	badgeRepo := Repository.NewDonorBadgeRepository(db)
	emergencyRepo := Repository.NewEmergencyRequestRepository(db)
	donorBloodReqRepo := Repository.NewDonorBloodRequestRepository(db)

	// --- Usecases ---
	badgeUsecase := Usecase.NewDonorBadgeUsecase(badgeRepo)
	campaignUsecase := Usecase.NewCampaignUsecase(campaignRepo)
	donationUsecase := Usecase.NewDonationUsecase(donationRepo, campaignRepo)
	labUsecase := Usecase.NewLabUsecase(labRepo, badgeUsecase)
	inventoryUsecase := Usecase.NewBloodInventoryUsecase(inventoryRepo)
	pdfService := Usecase.NewPDFGeneratorService("./uploads")
	hospitalUsecase := Usecase.NewHospitalUsecase(hospitalRepo, pdfService, userRepo)
	donorBloodReqUsecase := Usecase.NewDonorBloodRequestUsecase(donorBloodReqRepo)

	bloodReqRepo := Repository.NewBloodRequestRepository(db)
	
	// Start background jobs after all dependencies are initialized
	go Jobs.StartExpirationJob(inventoryUsecase, bloodReqRepo)

	campaignAnalyticsUsecase := Usecase.NewCampaignAnalyticsUsecase(campaignAnalyticsRepo)
	collectorAnalyticsUsecase := Usecase.NewCollectorAnalyticsUsecase(collectorAnalyticsRepo)
	labAnalyticsUsecase := Usecase.NewLabAnalyticsUsecase(labAnalyticsRepo)
	adminAnalyticsUsecase := Usecase.NewAdminAnalyticsUsecase(adminAnalyticsRepo)
	emergencyUsecase := Usecase.NewEmergencyRequestUsecase(emergencyRepo, inventoryRepo, hospitalRepo, bloodReqRepo, userRepo, profileRepo)
	bloodReqUsecase := Usecase.NewBloodRequestUsecase(bloodReqRepo, hospitalRepo, inventoryRepo, emergencyUsecase)

	// --- Controllers ---
	campaignController := controller.NewCampaignController(campaignUsecase)
	donationController := controller.NewDonationController(donationUsecase, userUseCase)
	labController := controller.NewLabController(labUsecase, userUseCase)
	inventoryController := controller.NewBloodInventoryController(inventoryUsecase)
	hospitalController := controller.NewHospitalController(hospitalUsecase)
	bloodReqController := controller.NewBloodRequestController(bloodReqUsecase)
	campaignAnalyticsController := controller.NewCampaignAnalyticsController(campaignAnalyticsUsecase)
	collectorAnalyticsController := controller.NewCollectorAnalyticsController(collectorAnalyticsUsecase)
	labAnalyticsController := controller.NewLabAnalyticsController(labAnalyticsUsecase)
	adminAnalyticsController := controller.NewAdminAnalyticsController(adminAnalyticsUsecase)
	badgeController := controller.NewDonorBadgeController(badgeUsecase, userUseCase)
	emergencyController := controller.NewEmergencyRequestController(emergencyUsecase)
	donorBloodReqController := controller.NewDonorBloodRequestController(donorBloodReqUsecase, userUseCase)

	// 5. Initialize Router
	r := router.SetupRouter(
		userController,
		jwtService,
		campaignController,
		donationController,
		labController,
		inventoryController,
		hospitalController,
		bloodReqController,
		campaignAnalyticsController,
		collectorAnalyticsController,
		labAnalyticsController,
		adminAnalyticsController,
		badgeController,
		emergencyController,
		donorBloodReqController,
	)

	// 7. Start the Server
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
