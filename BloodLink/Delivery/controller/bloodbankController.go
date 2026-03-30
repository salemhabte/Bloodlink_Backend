package controller

import (
	"bloodlink/Domain"
	"bloodlink/Usecase"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ==============================
//      CAMPAIGN CONTROLLER IMPLEMENTATION
// ==============================
type CampaignController struct {
    Usecase *Usecase.CampaignUsecase
}

func NewCampaignController(usecase *Usecase.CampaignUsecase) *CampaignController {
    return &CampaignController{Usecase: usecase}
}

func (c *CampaignController) CreateCampaign(ctx *gin.Context) {
    var input struct {
        Title     string    `json:"title" binding:"required"`
        Content   string    `json:"content" binding:"required"`
        Location  string    `json:"location" binding:"required"`
        StartDate time.Time `json:"start_date" binding:"required"`
        EndDate   time.Time `json:"end_date" binding:"required"`
    }

    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    campaign := &Domain.Campaign{
        Title:     input.Title,
        Content:   input.Content,
        Location:  input.Location,
        StartDate: input.StartDate,
        EndDate:   input.EndDate,
    }

    if err := c.Usecase.CreateCampaign(campaign); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusCreated, campaign)
}

func (c *CampaignController) GetAllCampaigns(ctx *gin.Context) {
    campaigns, err := c.Usecase.GetAllCampaigns()
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, campaigns)
}
func (c *CampaignController) GetCampaignByID(ctx *gin.Context) {
	id := ctx.Param("id")

	campaign, err := c.Usecase.GetCampaignByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, campaign)
}

func (c *CampaignController) UpdateCampaign(ctx *gin.Context) {
	id := ctx.Param("id")

	var input struct {
		Title     string    `json:"title"`
		Content   string    `json:"content"`
		Location  string    `json:"location"`
		StartDate time.Time `json:"start_date"`
		EndDate   time.Time `json:"end_date"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	campaign := &Domain.Campaign{
		CampaignID: id,
		Title:      input.Title,
		Content:    input.Content,
		Location:   input.Location,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
	}

	if err := c.Usecase.UpdateCampaign(campaign); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Campaign updated successfully"})
}

func (c *CampaignController) DeleteCampaign(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.Usecase.DeleteCampaign(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Campaign deleted successfully"})
}
// Search campaigns by location
func (c *CampaignController) GetCampaignsByLocation(ctx *gin.Context) {

	location := ctx.Query("location")

	campaigns, err := c.Usecase.GetCampaignsByLocation(location)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, campaigns)
}

// DonationController handles HTTP requests for the blood collector
type DonationController struct {
	usecase *Usecase.DonationUsecase
}

// Constructor
func NewDonationController(usecase *Usecase.DonationUsecase) *DonationController {
	return &DonationController{usecase: usecase}
}

// SearchDonor handles GET /bloodcollector/donor?email=
func (c *DonationController) SearchDonor(ctx *gin.Context) {
	// Get query from URL
	query := ctx.Query("q") // q is email or phone

	// Trim spaces to avoid hidden character issues
	query = strings.TrimSpace(query)

	// Debug: print what we actually received
	fmt.Printf("SearchDonor query received: '%s'\n", query)

	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Search value is required"})
		return
	}

	// Call usecase
	donor, err := c.usecase.SearchDonor(query)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Donor not found"})
		return
	}

	ctx.JSON(http.StatusOK, donor)
}

// CreateDonation handles POST /bloodcollector/donation
func (c *DonationController) CreateDonation(ctx *gin.Context) {
	var record Domain.DonationRecord
	if err := ctx.ShouldBindJSON(&record); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	 fmt.Printf("Inserting donation: %+v", record)

	if err := c.usecase.CreateDonation(&record); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, record)
}
func (c *DonationController) GetPendingDonors(ctx *gin.Context) {

	donors, err := c.usecase.GetPendingDonors()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, donors)
}
func (c *DonationController) GetDonorByID(ctx *gin.Context) {

	id := ctx.Param("id")

	donor, err := c.usecase.GetPendingDonorByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Donor not found"})
		return
	}

	ctx.JSON(http.StatusOK, donor)
}
func (c *DonationController) SearchPendingDonor(ctx *gin.Context) {

	query := ctx.Query("q") // ?q=email_or_phone

	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "query parameter 'q' is required",
		})
		return
	}

	donor, err := c.usecase.SearchPendingDonor(query)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, donor)
}
// UpdateDonationStatus handles PUT /bloodcollector/donation/:id/status
func (c *DonationController) UpdateDonationStatus(ctx *gin.Context) {
	donationID := ctx.Param("id")
	var body struct {
		Status string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.usecase.UpdateDonationStatus(donationID, body.Status); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}


func (c *DonationController) GetDonationByID(ctx *gin.Context) {

	id := ctx.Param("id")

	donation, err := c.usecase.GetDonationByID(id)

	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Donation not found"})
		return
	}

	ctx.JSON(http.StatusOK, donation)
}
func (c *DonationController) GetAllDonations(ctx *gin.Context) {

	donations, err := c.usecase.GetAllDonations()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, donations)
}
func (c *DonationController) UpdateDonation(ctx *gin.Context) {

	id := ctx.Param("id")

	var record Domain.DonationRecord

	if err := ctx.ShouldBindJSON(&record); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record.DonationID = id

	if err := c.usecase.UpdateDonation(&record); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Donation updated successfully"})
}

// LabController handles lab technician requests
type LabController struct {
	usecase *Usecase.LabUsecase
}

func NewLabController(usecase *Usecase.LabUsecase) *LabController {
	return &LabController{usecase: usecase}
}

// POST /api/lab/tests
func (c *LabController) SubmitTestResult(ctx *gin.Context) {
	var input Domain.DonorTestResult

	// 1. Bind request
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	// 2. Get logged-in lab technician from JWT
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	input.TestedBy = userID.(string)

	// 3. Check if test already exists
	existing, _ := c.usecase.GetTestResult(input.DonationID)
	if existing != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "a test for this donation already exists"})
		return
	}

	// 4. Process test
	err := c.usecase.ProcessTestResult(&input)
	if err != nil {
		// Suggestion logic (your nice feature)
		if strings.HasPrefix(err.Error(), "⚠ Suggestion:") {
			ctx.JSON(http.StatusBadRequest, gin.H{"warning": err.Error()})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 5. Success
	ctx.JSON(http.StatusOK, gin.H{
		"message": "test result processed successfully",
	})
}

// GET /api/lab/test-result/:donation_id
func (c *LabController) GetTestResult(ctx *gin.Context) {
	donationID := ctx.Param("donation_id")
	if donationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "donation_id is required"})
		return
	}

	result, err := c.usecase.GetTestResult(donationID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "test result not found"})
		return
	}
	// Convert boolean → readable text
	response := gin.H{
		"test_id": result.TestID,
		"donation_id": result.DonationID,
		"donor_id": result.DonorID,
		"tested_by": result.TestedBy,

		"hiv_result":        result.HIVResult,
		"hepatitis_result":  result.HepatitisResult,
		"syphilis_result":   result.SyphilisResult,

		"blood_type": result.BloodType,
		"overall_status": result.OverallStatus,
		"created_at": result.CreatedAt,
	}

	ctx.JSON(http.StatusOK, response)
}


func (c *LabController) GetPendingTests(ctx *gin.Context) {
	data, err := c.usecase.GetPendingDonations()
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, data)
}
func (c *LabController) GetHistory(ctx *gin.Context) {
	data, err := c.usecase.GetAllTestResults()
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, data)
}
func (c *LabController) FilterTests(ctx *gin.Context) {
	status := ctx.Query("status")

	data, err := c.usecase.GetTestResultsByStatus(status)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, data)
}
func (c *LabController) UpdateTest(ctx *gin.Context) {
	donationID := ctx.Param("donation_id")
	var input Domain.DonorTestResult

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(400, gin.H{"error": "invalid input"})
		return
	}
	input.DonationID = donationID

	err := c.usecase.UpdateTestResult(&input)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"message": "updated"})
}
func (c *LabController) RejectBlood(ctx *gin.Context) {
	id := ctx.Param("donation_id")

	err := c.usecase.RejectBlood(id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"message": "blood rejected"})
}
func (c *LabController) GetDonation(ctx *gin.Context) {
	donationID := ctx.Param("donation_id")

	if donationID == "" {
		ctx.JSON(400, gin.H{"error": "donation_id is required"})
		return
	}

	data, err := c.usecase.GetDonation(donationID)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "donation not found"})
		return
	}

	ctx.JSON(200, data)
}