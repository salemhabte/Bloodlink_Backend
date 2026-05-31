package controller

import (
	"bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"
	"bloodlink/Usecase"
	"fmt"
	"net/http"
	"strings"
	"time"

	"encoding/csv"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

// ==============================
//
//	CAMPAIGN CONTROLLER IMPLEMENTATION
//
// ==============================
type CampaignController struct {
	Usecase    *Usecase.CampaignUsecase
	auditLogger *Usecase.AuditLogUsecase
}

func NewCampaignController(usecase *Usecase.CampaignUsecase) *CampaignController {
	return &CampaignController{Usecase: usecase}
}

func (c *CampaignController) SetAuditLogger(logger *Usecase.AuditLogUsecase) {
	c.auditLogger = logger
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

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "CREATE_CAMPAIGN", "campaigns", campaign.CampaignID, "Created campaign: "+campaign.Title)
	}

	ctx.JSON(http.StatusCreated, campaign)
}

// GET /api/campaigns (Public/Donor) - Live Only
func (c *CampaignController) GetAllCampaigns(ctx *gin.Context) {
	filter := Domain.CampaignFilter{
		Title:     ctx.Query("title"),
		Location:  ctx.Query("location"),
		StartDate: ctx.Query("start_date"),
		EndDate:   ctx.Query("end_date"),
	}

	response, err := c.Usecase.GetAllCampaigns(filter, true) // LIVE ONLY
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// GET /api/bloodbankadmin/campaigns - All Campaigns
func (c *CampaignController) GetAllAdminCampaigns(ctx *gin.Context) {
	filter := Domain.CampaignFilter{
		Title:     ctx.Query("title"),
		Location:  ctx.Query("location"),
		StartDate: ctx.Query("start_date"),
		EndDate:   ctx.Query("end_date"),
		LiveOnly:  ctx.Query("live_only") == "true",
	}

	response, err := c.Usecase.GetAllCampaigns(filter, false) // ALL
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
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

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "UPDATE_CAMPAIGN", "campaigns", id, "Updated campaign: "+campaign.Title)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Campaign updated successfully"})
}

func (c *CampaignController) DeleteCampaign(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.Usecase.DeleteCampaign(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "DELETE_CAMPAIGN", "campaigns", id, "Deleted campaign")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Campaign deleted successfully"})
}

// DonationController handles HTTP requests for the blood collector
type DonationController struct {
	usecase     *Usecase.DonationUsecase
	userUsecase domainInterface.IUserUseCase
}

// Constructor
func NewDonationController(usecase *Usecase.DonationUsecase, userUsecase domainInterface.IUserUseCase) *DonationController {
	return &DonationController{usecase: usecase, userUsecase: userUsecase}
}

// CreateDonation handles POST /bloodcollector/donation
func (c *DonationController) CreateDonation(ctx *gin.Context) {
	var record Domain.DonationRecord

	if err := ctx.ShouldBindJSON(&record); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Inject collector_id from JWT (never trust request body)
	collectorID := ctx.GetString("userID")
	record.CollectedBy = collectorID

	fmt.Printf("Inserting donation: %+v", record)

	if err := c.usecase.CreateDonation(&record); err != nil {
		// Check if it's a suggestion warning
		if strings.HasPrefix(err.Error(), "⚠ Suggestion:") {
			ctx.JSON(http.StatusBadRequest, gin.H{"warning": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, record)
}
func (c *DonationController) GetDonorByID(ctx *gin.Context) {

	id := ctx.Param("id")

	donor, err := c.userUsecase.GetEligibleDonorByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, donor)
}

// UpdateDonationStatus handles PUT /bloodcollector/donation/:id/status
func (c *DonationController) UpdateDonationStatus(ctx *gin.Context) {

	donationID := ctx.Param("id")

	var body struct {
		Status          string `json:"status"`
		RejectionReason string `json:"rejection_reason"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//  GET LOGGED IN COLLECTOR
	collectorID := ctx.GetString("userID")

	// SECURE CALL
	if err := c.usecase.UpdateDonationStatus(donationID, body.Status, body.RejectionReason, collectorID); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

// UpdateDonation medical info
func (c *DonationController) UpdateDonation(ctx *gin.Context) {
	id := ctx.Param("id")
	var record Domain.DonationRecord
	if err := ctx.ShouldBindJSON(&record); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	record.DonationID = id
	collectorID := ctx.GetString("userID")
	record.CollectedBy = collectorID

	if err := c.usecase.UpdateDonation(&record); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Donation updated successfully"})
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

func (c *DonationController) GetAllDonationsByDonor(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	donorID, err := c.userUsecase.GetDonorIDByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "could not identify donor profile: " + err.Error()})
		return
	}

	donations, err := c.usecase.GetAllDonationsByDonor(donorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, donations)
}

func (c *DonationController) GetAllDonations(ctx *gin.Context) {

	filter := Domain.DonationFilter{
		DonorID:        ctx.Query("donor_id"),
		CollectorID:    ctx.Query("collector_id"),
		Status:         ctx.Query("status"),
		DonationNumber: ctx.Query("donation_number"),
		StartDate:      ctx.Query("start_date"),
		EndDate:        ctx.Query("end_date"),
	}

	data, err := c.usecase.GetAllDonations(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := Domain.DonationListResponse{
		Total:     len(data),
		Donations: data,
	}
	for _, d := range data {
		status := strings.ToUpper(d.Status)
		if status == "APPROVED" {
			resp.Approved++
		} else if status == "REJECTED_TEMPORARY" {
			resp.TemporarilyRejected++
		}
	}

	ctx.JSON(http.StatusOK, resp)
}
func (c *DonationController) GetMyDonations(ctx *gin.Context) {

	collectorID := ctx.GetString("userID")

	filter := Domain.DonationFilter{
		Status:         ctx.Query("status"),
		DonationNumber: ctx.Query("donation_number"),
		StartDate:      ctx.Query("start_date"),
		EndDate:        ctx.Query("end_date"),
	}

	data, err := c.usecase.GetMyDonations(collectorID, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := Domain.DonationListResponse{
		Total:     len(data),
		Donations: data,
	}
	for _, d := range data {
		status := strings.ToUpper(d.Status)
		if status == "APPROVED" {
			resp.Approved++
		} else if status == "REJECTED_TEMPORARY" {
			resp.TemporarilyRejected++
		}
	}

	ctx.JSON(http.StatusOK, resp)
}

// LabController handles lab technician requests
type LabController struct {
	usecase     *Usecase.LabUsecase
	userUsecase domainInterface.IUserUseCase
}

func NewLabController(usecase *Usecase.LabUsecase, userUsecase domainInterface.IUserUseCase) *LabController {
	return &LabController{usecase: usecase, userUsecase: userUsecase}
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
		"message":     "test result processed successfully",
		"test_result": input,
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

	ctx.JSON(http.StatusOK, result)
}

func (c *LabController) GetPendingDonations(ctx *gin.Context) {
	data, err := c.usecase.GetPendingDonations()
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

	// 🔐 GET LOGGED-IN LAB TECH ID (IMPORTANT FIX)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// FORCE ownership from JWT (DO NOT TRUST BODY)
	input.TestedBy = userID.(string)

	// CALL USECASE
	err := c.usecase.UpdateTestResult(&input, userID.(string))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"message": "updated"})
}
func (c *LabController) RejectBlood(ctx *gin.Context) {

	id := ctx.Param("donation_id")

	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	err := c.usecase.RejectBlood(id, userID.(string))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"message": "blood rejected"})
}

func (c *LabController) GetLatestTestResultByDonor(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	donorID, err := c.userUsecase.GetDonorIDByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "could not identify donor profile: " + err.Error()})
		return
	}

	result, err := c.usecase.GetLatestTestResultByDonor(donorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "no test results found for this donor"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
func (c *LabController) GetPendingDonation(ctx *gin.Context) {
	donationID := ctx.Param("donation_id")

	if donationID == "" {
		ctx.JSON(400, gin.H{"error": "donation_id is required"})
		return
	}

	data, err := c.usecase.GetPendingDonation(donationID)
	if err != nil {
		ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, data)
}
func (c *LabController) GetMyTests(ctx *gin.Context) {

	// get logged-in lab tech
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	labTechID := userID.(string)

	// filters
	filter := Domain.TestFilter{
		LabTechID:       labTechID,
		OverallStatus:   strings.ToUpper(strings.TrimSpace(ctx.Query("overall_status"))),
		BloodType:       normalizeBloodType(ctx.Query("blood_type")),
		ComponentType:   strings.ToUpper(strings.TrimSpace(ctx.Query("component_type"))),
		StorageLocation: ctx.Query("storage_location"),
		DonationNumber:  ctx.Query("donation_number"),
		StartDate:       ctx.Query("start_date"),
		EndDate:         ctx.Query("end_date"),
	}

	// call usecase
	data, err := c.usecase.GetMyTestResultsFiltered(filter)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, c.wrapTestResults(data))
}
// ✅ FIXED: safe normalization (DO NOT touch encoding)
func normalizeBloodType(input string) string {
	return strings.ToUpper(strings.TrimSpace(input))
}
func (c *LabController) GetAllTestHistory(ctx *gin.Context) {

	filter := Domain.TestFilter{
		OverallStatus:   strings.ToUpper(strings.TrimSpace(ctx.Query("overall_status"))),
		BloodType:       normalizeBloodType(ctx.Query("blood_type")),
		ComponentType:   strings.ToUpper(strings.TrimSpace(ctx.Query("component_type"))),
		StorageLocation: ctx.Query("storage_location"),
		DonationNumber:  ctx.Query("donation_number"),
		StartDate:       ctx.Query("start_date"),
		EndDate:         ctx.Query("end_date"),
	}

	data, err := c.usecase.GetAllTestsFiltered(filter)

	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, c.wrapTestResults(data))
}

func (c *LabController) wrapTestResults(tests []Domain.DonorTestResult) Domain.TestResultListResponse {
	resp := Domain.TestResultListResponse{
		Total: len(tests),
		Tests: tests,
	}
	for _, t := range tests {
		status := strings.ToUpper(t.OverallStatus)
		if status == "CLEARED" {
			resp.Cleared++
		} else if status == "TEMPORARILY_DEFERRED" {
			resp.TemporarilyDeferred++
		} else if status == "PERMANENTLY_DEFERRED" {
			resp.PermanentlyDeferred++
		}
	}
	return resp
}

// BloodInventoryController
type BloodInventoryController struct {
	usecase     *Usecase.BloodInventoryUsecase
	auditLogger *Usecase.AuditLogUsecase
}

func NewBloodInventoryController(u *Usecase.BloodInventoryUsecase) *BloodInventoryController {
	return &BloodInventoryController{usecase: u}
}

func (c *BloodInventoryController) SetAuditLogger(logger *Usecase.AuditLogUsecase) {
	c.auditLogger = logger
}

// 🔹 GET /inventory
func (c *BloodInventoryController) GetAll(ctx *gin.Context) {
	vol, _ := strconv.Atoi(ctx.Query("quantity"))
	bloodType := ctx.Query("blood_type")
	if bloodType != "" {
		bloodType = strings.ReplaceAll(bloodType, " ", "+")
	}

	filter := Domain.BloodUnitFilter{
		BloodType:     bloodType,
		ComponentType: ctx.Query("component_type"),
		Status:        ctx.Query("status"),
		StartDate:     ctx.Query("start_date"),
		EndDate:       ctx.Query("end_date"),
		Quantity:      vol,
		NearExpired:   ctx.Query("near_expired") == "true",
		UnitNumber:    ctx.Query("unit_number"),
	}

	if (filter.StartDate != "" && filter.EndDate == "") || (filter.StartDate == "" && filter.EndDate != "") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Both start_date and end_date are required"})
		return
	}

	res, err := c.usecase.GetAllUnits(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (c *BloodInventoryController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")

	unit, err := c.usecase.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Blood unit not found"})
		return
	}

	ctx.JSON(http.StatusOK, unit)
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



// 🔹 PUT /inventory/:id/used
func (c *BloodInventoryController) MarkUsed(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.usecase.MarkUnitAsUsed(id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "MARK_UNIT_USED", "blood_units", id, "Marked blood unit as used")
	}

	ctx.JSON(200, gin.H{"message": "Blood unit marked as USED"})
}

// 🔹 DELETE /inventory/:id
func (c *BloodInventoryController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.usecase.DeleteUnit(id)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "DELETE_UNIT", "blood_units", id, "Deleted blood unit")
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

func (c *BloodInventoryController) GetReservedByHospital(ctx *gin.Context) {
	hospitalID := ctx.Param("hospital_id")

	units, err := c.usecase.GetReservedUnitsByHospital(hospitalID)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, units)
}

func (c *BloodInventoryController) GetReservedByRequest(ctx *gin.Context) {
	requestID := ctx.Param("request_id")

	units, err := c.usecase.GetReservedUnitsByRequestID(requestID)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, units)
}

// 🔹 POST /inventory/:id/convert-cryo
func (c *BloodInventoryController) ConvertPlasma(ctx *gin.Context) {
	id := ctx.Param("id")
	
	var req Domain.ConvertCryoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.usecase.ConvertPlasmaToCryo(
		id,
		req.CryoprecipitateQuantity,
		req.CryoPoorPlasmaQuantity,
		req.CryoStorageLocation, req.CryoRackNumber, req.CryoShelfNumber, req.CryoPositionNumber,
		req.CryoPoorStorageLocation, req.CryoPoorRackNumber, req.CryoPoorShelfNumber, req.CryoPoorPositionNumber,
	)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Plasma converted to Cryoprecipitate successfully"})
}

func (c *BloodInventoryController) ExportCSV(ctx *gin.Context) {
	vol, _ := strconv.Atoi(ctx.Query("quantity"))
	bloodType := ctx.Query("blood_type")
	if bloodType != "" {
		bloodType = strings.ReplaceAll(bloodType, " ", "+")
	}

	filter := Domain.BloodUnitFilter{
		BloodType:     bloodType,
		ComponentType: ctx.Query("component_type"),
		Status:        ctx.Query("status"),
		StartDate:     ctx.Query("start_date"),
		EndDate:       ctx.Query("end_date"),
		Quantity:        vol,
		NearExpired:   ctx.Query("near_expired") == "true",
	}

	res, _ := c.usecase.GetAllUnits(filter)
	units := res.Units

	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", `attachment; filename="blood_inventory.csv"`)
	ctx.Header("Content-Transfer-Encoding", "binary")

	writer := csv.NewWriter(ctx.Writer)
	defer writer.Flush()

	writer.Write([]string{
		"unit_number",
		"blood_type",
		"component_type",
		"quantity_ml",
		"collection_date",
		"expiration_date",
		"status",
		"storage_location",
		"rack_number",
		"shelf_number",
		"position_number",
	})

	for _, u := range units {
		writer.Write([]string{
			u.UnitNumber,
			u.BloodType,
			u.ComponentType,
			strconv.Itoa(u.QuantityML),
			u.CollectionDate.Format("2006-01-02"),
			u.ExpirationDate.Format("2006-01-02"),
			u.Status,
			u.StorageLocation,
			u.RackNumber,
			u.ShelfNumber,
			u.PositionNumber,
		})
	}
}
func (c *BloodInventoryController) ExportPDF(ctx *gin.Context) {
	vol, _ := strconv.Atoi(ctx.Query("quantity"))
	bloodType := ctx.Query("blood_type")
	if bloodType != "" {
		bloodType = strings.ReplaceAll(bloodType, " ", "+")
	}

	filter := Domain.BloodUnitFilter{
		BloodType:     bloodType,
		ComponentType: ctx.Query("component_type"),
		Status:        ctx.Query("status"),
		StartDate:     ctx.Query("start_date"),
		EndDate:       ctx.Query("end_date"),
		Quantity:      vol,
		NearExpired:   ctx.Query("near_expired") == "true",
		UnitNumber:    ctx.Query("unit_number"),
	}

	res, err := c.usecase.GetAllUnits(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	units := res.Units

	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(280, 10, "Blood Inventory Report")
	pdf.Ln(12)

	// 🔹 Header Row
	pdf.SetFont("Arial", "B", 8)

	colID := 40.0
	colType := 20.0
	colComp := 30.0
	colQty := 15.0
	colDate := 20.0
	colStatus := 20.0
	colStore := 40.0
	colRack := 15.0
	colShelf := 15.0
	colPos := 15.0

	pdf.CellFormat(colID, 10, "Unit Number", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colType, 10, "Blood Type", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colComp, 10, "Component", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colQty, 10, "Qty (ml)", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colDate, 10, "Expiry", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colStatus, 10, "Status", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colStore, 10, "Location", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colRack, 10, "Rack", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colShelf, 10, "Shelf", "1", 0, "C", false, 0, "")
	pdf.CellFormat(colPos, 10, "Pos", "1", 0, "C", false, 0, "")
	pdf.Ln(-1)

	// 🔹 Data Rows
	pdf.SetFont("Arial", "", 7)

	for _, u := range units {
		if u.Status == "EXPIRED" {
			pdf.SetTextColor(255, 0, 0)
		} else {
			pdf.SetTextColor(0, 0, 0)
		}

		pdf.CellFormat(colID, 8, u.UnitNumber, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colType, 8, u.BloodType, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colComp, 8, u.ComponentType, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colQty, 8, strconv.Itoa(u.QuantityML), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colDate, 8, u.ExpirationDate.Format("2006-01-02"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colStatus, 8, u.Status, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colStore, 8, u.StorageLocation, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colRack, 8, u.RackNumber, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colShelf, 8, u.ShelfNumber, "1", 0, "C", false, 0, "")
		pdf.CellFormat(colPos, 8, u.PositionNumber, "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.SetTextColor(0, 0, 0)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", `attachment; filename="blood_inventory.pdf"`)
	ctx.Header("Content-Transfer-Encoding", "binary")
	pdf.Output(ctx.Writer)
}


func (c *BloodInventoryController) wrapInventoryResults(units []Domain.BloodUnit) Domain.InventoryListResponse {
	resp := Domain.InventoryListResponse{
		Total:               len(units),
		ByBloodType:         make(map[string]int),
		ByComponentType:     make(map[string]int),
		ByBloodAndComponent: make(map[string]int),
		Units:               units,
	}

	now := time.Now()
	nearExpiryLimit := now.AddDate(0, 0, 7)

	for _, u := range units {
		resp.ByBloodType[u.BloodType]++
		if u.ComponentType != "" {
			resp.ByComponentType[u.ComponentType]++
			resp.ByBloodAndComponent[u.BloodType+"_"+u.ComponentType]++
		}

		status := strings.ToUpper(u.Status)
		switch status {
		case "AVAILABLE":
			resp.Available++
			if u.ExpirationDate.Before(nearExpiryLimit) && u.ExpirationDate.After(now) {
				resp.NearExpired++
			}
		case "RESERVED":
			resp.Reserved++
		case "USED":
			resp.Used++
		case "EXPIRED":
			resp.Expired++
		}
	}
	return resp
}
