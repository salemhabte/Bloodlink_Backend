package controller

import (
	"bloodlink/Domain"
	"bloodlink/Usecase"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type DonorBloodRequestController struct {
	usecase *Usecase.DonorBloodRequestUsecase
	auditUsecase *Usecase.AuditLogUsecase
}

func NewDonorBloodRequestController(
	u *Usecase.DonorBloodRequestUsecase,
	au *Usecase.AuditLogUsecase,
) *DonorBloodRequestController {
	return &DonorBloodRequestController{
		usecase: u,
		auditUsecase: au,
	}
}

////////////////////////
// CREATE REQUEST (DONOR)
////////////////////////

func (c *DonorBloodRequestController) CreateRequest(ctx *gin.Context) {

	userID := ctx.GetString("userID")

	var req struct {
		QuantityML      int    `json:"quantity_ml"`
		Reason          string `json:"reason"`
		HospitalName    string `json:"hospital_name"`
		HospitalAddress string `json:"hospital_address"`
		HospitalPhone   string `json:"hospital_phone"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.usecase.CreateRequest(
		userID,
		req.QuantityML,
		req.Reason,
		req.HospitalName,
		req.HospitalAddress,
		req.HospitalPhone,
	)

	if err != nil {
		// Surface the "not enough donations" message as a clear 403
		if strings.Contains(err.Error(), "at least perform one successful donation") {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Blood request created successfully",
	})
}

////////////////////////
// GET MY REQUESTS (DONOR)
////////////////////////

func (c *DonorBloodRequestController) GetMyRequests(ctx *gin.Context) {

	userID := ctx.GetString("userID")

	filter := Domain.DonorBloodRequestFilter{
		StartDate: ctx.Query("start_date"),
		EndDate:   ctx.Query("end_date"),
		Status:    ctx.Query("status"),
	}

	data, err := c.usecase.GetMyRequests(userID, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Just in case, ensure it's not null but [] if empty
	if data == nil {
		data = []Domain.DonorBloodRequest{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "My requests fetched successfully",
		"data":    data,
	})
}

////////////////////////
// GET ALL REQUESTS (ADMIN) — simple unfiltered version
////////////////////////

func (c *DonorBloodRequestController) GetAllRequests(ctx *gin.Context) {

	data, err := c.usecase.GetAllRequests()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "All requests fetched successfully",
		"data":    data,
	})
}

////////////////////////
// GET ALL REQUESTS — ADMIN FILTERED (sorted by successful donations DESC)
// Query params: start_date, end_date, blood_type, status
////////////////////////

func (c *DonorBloodRequestController) GetAllAdminRequests(ctx *gin.Context) {

	filter := Domain.DonorBloodRequestFilter{
		StartDate: ctx.Query("start_date"),
		EndDate:   ctx.Query("end_date"),
		BloodType: ctx.Query("blood_type"),
		Status:    ctx.Query("status"),
	}

	data, err := c.usecase.GetAllAdminRequests(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Admin donor requests fetched successfully",
		"data":    data,
	})
}

////////////////////////
// APPROVE (ADMIN)
// Returns the result message from the usecase:
//   "no enough blood in the inventory" → 200 but status is REJECTED
//   "partially approved"               → 200 + PARTIALLY APPROVED
//   "fully approved"                   → 200 + APPROVED
////////////////////////

func (c *DonorBloodRequestController) ApproveRequest(ctx *gin.Context) {

	id := ctx.Param("id")
	oldReq, _ := c.usecase.GetRequestByID(id)
	oldStatus := "N/A"
	targetName := "Donor Blood Request"
	if oldReq != nil {
		oldStatus = oldReq.Status
		targetName = "Donor Request for " + oldReq.HospitalName
	}

	updatedReq, message, err := c.usecase.ApproveRequest(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.auditUsecase.Log(ctx.Request.Context(), userID, "APPROVE_DONOR_BLOOD_REQUEST", "DONOR_BLOOD_REQUEST", id, targetName, oldStatus, updatedReq.Status)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": message,
		"data":    updatedReq,
	})
}

////////////////////////
// REJECT (ADMIN)
////////////////////////

func (c *DonorBloodRequestController) RejectRequest(ctx *gin.Context) {

	id := ctx.Param("id")
	oldReq, _ := c.usecase.GetRequestByID(id)
	oldStatus := "N/A"
	targetName := "Donor Blood Request"
	if oldReq != nil {
		oldStatus = oldReq.Status
		targetName = "Donor Request for " + oldReq.HospitalName
	}

	updatedReq, err := c.usecase.RejectRequest(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.auditUsecase.Log(ctx.Request.Context(), userID, "REJECT_DONOR_BLOOD_REQUEST", "DONOR_BLOOD_REQUEST", id, targetName, oldStatus, "REJECTED")
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Request rejected successfully",
		"data":    updatedReq,
	})
}

////////////////////////
// FULFILL (ADMIN)
// Transitions:
//   APPROVED          → FULFILLED
//   PARTIALLY APPROVED → PARTIALLY FULFILLED
////////////////////////

func (c *DonorBloodRequestController) FulfillRequest(ctx *gin.Context) {

	id := ctx.Param("id")

	oldReq, _ := c.usecase.GetRequestByID(id)
	oldStatus := "N/A"
	targetName := "Donor Blood Request"
	if oldReq != nil {
		oldStatus = oldReq.Status
		targetName = "Donor Request for " + oldReq.HospitalName
	}

	updatedReq, err := c.usecase.FulfillRequest(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.auditUsecase.Log(ctx.Request.Context(), userID, "FULFILL_DONOR_BLOOD_REQUEST", "DONOR_BLOOD_REQUEST", id, targetName, oldStatus, "FULFILLED")
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Request fulfilled successfully",
		"data":    updatedReq,
	})
}

////////////////////////
// EXPIRE STALE RESERVATIONS (ADMIN / CRON)
// Call this endpoint periodically (or via a cron job) to release
// any blood units that have been reserved for > 24 hours without
// the admin clicking "Fulfilled".
////////////////////////

func (c *DonorBloodRequestController) ExpireStaleRequests(ctx *gin.Context) {

	err := c.usecase.ExpireStaleRequests()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.auditUsecase.Log(ctx.Request.Context(), userID, "EXPIRE_STALE_DONOR_REQUESTS", "SYSTEM", "", "Batch Operation", "N/A", "EXPIRED")
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Stale reservations expired successfully",
	})
}