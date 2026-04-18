package controller

import (
	"bloodlink/Domain"
	"bloodlink/Usecase"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"strconv"
	"github.com/jung-kurt/gofpdf"
	"encoding/csv"
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

//BloodInventoryController
type BloodInventoryController struct {
	usecase *Usecase.BloodInventoryUsecase
}

func NewBloodInventoryController(u *Usecase.BloodInventoryUsecase) *BloodInventoryController {
	return &BloodInventoryController{usecase: u}
}

// 🔹 GET /inventory
func (c *BloodInventoryController) GetAll(ctx *gin.Context) {
	data, err := c.usecase.GetAllUnits()
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, data)
}

// 🔹 GET /inventory/stats
func (c *BloodInventoryController) GetStats(ctx *gin.Context) {
	stats, err := c.usecase.GetStats()
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(200, stats)
}

// 🔹 PUT /inventory/:id/status
func (c *BloodInventoryController) UpdateStatus(ctx *gin.Context) {
	id := ctx.Param("id")

	var body struct {
		Status string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid body"})
		return
	}

	err := c.usecase.UpdateStatus(id, body.Status)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"message": "Status updated"})
}

// 🔹 DELETE /inventory/:id
func (c *BloodInventoryController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.usecase.DeleteUnit(id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"message": "Deleted successfully"})
}
func (c *BloodInventoryController) GetFullDetails(ctx *gin.Context) {

	id := ctx.Param("id")

	data, err := c.usecase.GetFullDetails(id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, data)
}
func (c *BloodInventoryController) ExportCSV(ctx *gin.Context) {

	units, _ := c.usecase.GetAllUnits()

	ctx.Header("Content-Disposition", "attachment; filename=blood_units.csv")
	ctx.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(ctx.Writer)
	defer writer.Flush()

	writer.Write([]string{
    "blood_unit_id",
    "donation_id",
    "blood_type",
    "volume_ml",
    "collection_date",
    "expiration_date",
    "status",
})

for _, u := range units {
    writer.Write([]string{
        u.BloodUnitID,
        u.DonationID,
        u.BloodType,
        strconv.Itoa(u.VolumeML),
        u.CollectionDate.Format("2006-01-02"),
        u.ExpirationDate.Format("2006-01-02"),
        u.Status,
    })
}
}
func (c *BloodInventoryController) ExportPDF(ctx *gin.Context) {

    units, err := c.usecase.GetAllUnits()
    if err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }

    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()

    // 🔹 Title
    pdf.SetFont("Arial", "B", 16)
    pdf.Cell(190, 10, "Blood Inventory Report")
    pdf.Ln(12)

    // 🔹 Header Row
    pdf.SetFont("Arial", "B", 9)

    col1 := 45.0
    col2 := 45.0
    col3 := 15.0
    col4 := 15.0
    col5 := 25.0
    col6 := 25.0
    col7 := 20.0

   pdf.CellFormat(col1, 10, "Blood Unit ID", "1", 0, "C", false, 0, "")
pdf.CellFormat(col2, 10, "Donation ID", "1", 0, "C", false, 0, "")
pdf.CellFormat(col3, 10, "Blood Type", "1", 0, "C", false, 0, "")
pdf.CellFormat(col4, 10, "Volume (ml)", "1", 0, "C", false, 0, "")
pdf.CellFormat(col5, 10, "Collection Date", "1", 0, "C", false, 0, "")
pdf.CellFormat(col6, 10, "Expiry Date", "1", 0, "C", false, 0, "")
pdf.CellFormat(col7, 10, "Unit Status", "1", 0, "C", false, 0, "")
pdf.Ln(-1)
    // 🔹 Data Rows
    pdf.SetFont("Arial", "", 8)

    for _, u := range units {

        if u.Status == "EXPIRED" {
            pdf.SetTextColor(255, 0, 0)
        } else {
            pdf.SetTextColor(0, 0, 0)
        }

        // Save current position
        x := pdf.GetX()
        y := pdf.GetY()

        lineHeight := 5.0
        maxHeight := lineHeight

        // --- Column 1 (Unit ID) ---
        pdf.SetXY(x, y)
        pdf.MultiCell(col1, lineHeight, u.BloodUnitID, "1", "L", false)
        h1 := pdf.GetY() - y

        // --- Column 2 (Donation ID) ---
        pdf.SetXY(x+col1, y)
        pdf.MultiCell(col2, lineHeight, u.DonationID, "1", "L", false)
        h2 := pdf.GetY() - y

        // Determine max height
        if h1 > maxHeight {
            maxHeight = h1
        }
        if h2 > maxHeight {
            maxHeight = h2
        }

        // --- Remaining columns ---
        pdf.SetXY(x+col1+col2, y)
        pdf.CellFormat(col3, maxHeight, u.BloodType, "1", 0, "C", false, 0, "")

        pdf.CellFormat(col4, maxHeight, strconv.Itoa(u.VolumeML), "1", 0, "C", false, 0, "")

        pdf.CellFormat(col5, maxHeight, u.CollectionDate.Format("2006-01-02"), "1", 0, "C", false, 0, "")

        pdf.CellFormat(col6, maxHeight, u.ExpirationDate.Format("2006-01-02"), "1", 0, "C", false, 0, "")

        pdf.CellFormat(col7, maxHeight, u.Status, "1", 0, "C", false, 0, "")

        // Move to next row
        pdf.Ln(maxHeight)
    }

    pdf.SetTextColor(0, 0, 0)

    ctx.Header("Content-Type", "application/pdf")
    ctx.Header("Content-Disposition", "attachment; filename=blood_inventory.pdf")

    err = pdf.Output(ctx.Writer)
    if err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }
}
func (c *BloodInventoryController) Filter(ctx *gin.Context) {

	unitID := ctx.Query("unit_id")
	bloodType := ctx.Query("blood_type")
	status := ctx.Query("status")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	data, err := c.usecase.FilterUnits(unitID, bloodType, status, startDate, endDate)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, data)
}