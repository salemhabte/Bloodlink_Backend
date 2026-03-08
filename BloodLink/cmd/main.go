package main

import (
	"bloodlink/Delivery/controller"
	"bloodlink/Delivery/router"
	"bloodlink/Repository"
	"bloodlink/Usecase"
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	// Remote MySQL DSN from your teammate
	dsn := "sql12819087:NtpQbxQu4J@tcp(sql12.freesqldatabase.com:3306)/sql12819087"

	// Connect to DB
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer db.Close()

	// Check DB connection
	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to database:", err)
	}

	log.Println("Connected to remote MySQL database!")
	// ===== CAMPAIGN SECTION =====
	// --- Repositories ---
	campaignRepo := Repository.NewCampaignRepository(db)

	// --- Usecases ---
	campaignUsecase := Usecase.NewCampaignUsecase(campaignRepo)

	// --- Controllers ---
	campaignController := controller.NewCampaignController(campaignUsecase)

	// --- Router ---
	r := router.SetupRouter(campaignController)

	log.Println("Server running on port 8080...")
	log.Fatal(r.Run(":8080"))
}