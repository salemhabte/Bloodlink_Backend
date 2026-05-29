package controller

import (
	"bloodlink/Domain"
	"bloodlink/Usecase"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type DonorBloodRequestController struct {
	usecase    *Usecase.DonorBloodRequestUsecase
	auditLogger *Usecase.AuditLogUsecase
}

func NewDonorBloodRequestController(
	u *Usecase.DonorBloodRequestUsecase,
) *DonorBloodRequestController {
	return &DonorBloodRequestController{
		usecase: u,
	}
}

func (c *DonorBloodRequestController) SetAuditLogger(logger *Usecase.AuditLogUsecase) {
	c.auditLogger = logger
}

////////////////////////
// CREATE REQUEST (DONOR)
////////////////////////

func (c *DonorBloodRequestController) CreateRequest(ctx *gin.Context) {

	userID := ctx.GetString("userID")

	var req struct {
		Units           int    `json:"units"`
		ComponentType   string `json:"component_type"`
		Reason          string `json:"reason"`
		HospitalName    string `json:"hospital_name"`
		HospitalAddress string `json:"hospital_address"`
		HospitalPhone   string `json:"hospital_phone"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := c.usecase.CreateRequest(
		userID,
		req.Units,
		req.ComponentType,
		req.Reason,
		req.HospitalName,
		req.HospitalAddress,
		req.HospitalPhone,
	)

	if err != nil {
		if strings.Contains(err.Error(), "top 10") || strings.Contains(err.Error(), "3 months") {
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Blood request created successfully",
		"data":    created,
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

	if (filter.StartDate != "" && filter.EndDate == "") || (filter.StartDate == "" && filter.EndDate != "") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Both start_date and end_date are required"})
		return
	}

	res, err := c.usecase.GetMyRequests(userID, filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Just in case, ensure it's not null but [] if empty
	if res.Requests == nil {
		res.Requests = []Domain.DonorBloodRequest{}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "My requests fetched successfully",
		"data":    res,
	})
}

////////////////////////
// GET ALL REQUESTS — ADMIN FILTERED (sorted by successful donations DESC)
// Query params: start_date, end_date, blood_type, status
////////////////////////

func (c *DonorBloodRequestController) GetAllAdminRequests(ctx *gin.Context) {
	bloodType := ctx.Query("blood_type")
	if bloodType != "" {
		bloodType = strings.ReplaceAll(bloodType, " ", "+")
	}

	filter := Domain.DonorBloodRequestFilter{
		StartDate: ctx.Query("start_date"),
		EndDate:   ctx.Query("end_date"),
		BloodType: bloodType,
		Status:    ctx.Query("status"),
	}

	if (filter.StartDate != "" && filter.EndDate == "") || (filter.StartDate == "" && filter.EndDate != "") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Both start_date and end_date are required"})
		return
	}

	res, err := c.usecase.GetAllAdminRequests(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Admin donor requests fetched successfully",
		"data":    res,
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

	updatedReq, message, err := c.usecase.ApproveRequest(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "APPROVE_DONOR_BLOOD_REQUEST", "donor_blood_requests", id, "Approved donor blood request")
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

	updatedReq, err := c.usecase.RejectRequest(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "REJECT_DONOR_BLOOD_REQUEST", "donor_blood_requests", id, "Rejected donor blood request")
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

	updatedReq, err := c.usecase.FulfillRequest(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "FULFILL_DONOR_BLOOD_REQUEST", "donor_blood_requests", id, "Fulfilled donor blood request")
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

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Stale reservations expired successfully",
	})
}