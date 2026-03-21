package main

import (
	"bloodlink/Delivery/controller"
	"bloodlink/Delivery/router"
	"bloodlink/Infrastructure"
	"bloodlink/Repository"
	"bloodlink/Usecase"

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

	// --- Usecases ---
	campaignUsecase := Usecase.NewCampaignUsecase(campaignRepo)
	donationUsecase := Usecase.NewDonationUsecase(donationRepo)

	// --- Controllers ---
	campaignController := controller.NewCampaignController(campaignUsecase)
	donationController := controller.NewDonationController(donationUsecase)

	// --- Hospital ---
	hospitalRepo := Repository.NewHospitalRepository(db)
	hospitalUsecase := Usecase.NewHospitalUsecase(hospitalRepo)
	hospitalController := controller.NewHospitalController(hospitalUsecase)

	// 5. Initialize Router
	r := router.SetupRouter(userController, jwtService, campaignController, donationController, hospitalController)

	// 7. Start the Server
	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
