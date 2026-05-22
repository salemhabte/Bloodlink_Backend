package controller

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"bloodlink/Usecase"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type EmergencyRequestController struct {
	Usecase    Interfaces.IEmergencyRequestUsecase
	auditLogger *Usecase.AuditLogUsecase
}

func NewEmergencyRequestController(u Interfaces.IEmergencyRequestUsecase) *EmergencyRequestController {
	return &EmergencyRequestController{Usecase: u}
}

func (c *EmergencyRequestController) SetAuditLogger(logger *Usecase.AuditLogUsecase) {
	c.auditLogger = logger
}

func (c *EmergencyRequestController) CreateManualEmergency(ctx *gin.Context) {
	// 1. Try to bind as a slice (bulk request)
	var dtos []Domain.CreateEmergencyRequestDTO
	if err := ctx.ShouldBindJSON(&dtos); err == nil {
		if err := c.Usecase.CreateManualEmergency(dtos); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if c.auditLogger != nil {
			adminID := ctx.GetString("userID")
			c.auditLogger.LogAction(adminID, "CREATE_MANUAL_EMERGENCY", "emergency_requests", "", "Created bulk manual emergency requests")
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": "Manual emergency request(s) created and published successfully"})
		return
	}

	// 2. Fallback to single object
	var dto Domain.CreateEmergencyRequestDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: expected a single emergency or a list of emergencies"})
		return
	}

	if err := c.Usecase.CreateManualEmergency([]Domain.CreateEmergencyRequestDTO{dto}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "CREATE_MANUAL_EMERGENCY", "emergency_requests", "", "Created manual emergency request")
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Manual emergency request created and published successfully"})
}

func (c *EmergencyRequestController) PublishEmergency(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.Usecase.PublishEmergency(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "PUBLISH_EMERGENCY", "emergency_requests", id, "Published emergency request")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Emergency request published successfully"})
}

func (c *EmergencyRequestController) RejectEmergency(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.Usecase.RejectEmergency(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if c.auditLogger != nil {
		adminID := ctx.GetString("userID")
		c.auditLogger.LogAction(adminID, "REJECT_EMERGENCY", "emergency_requests", id, "Rejected emergency request")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Emergency request rejected successfully"})
}

func (c *EmergencyRequestController) GetAllEmergencies(ctx *gin.Context) {
	bloodType := ctx.Query("blood_type")
	if bloodType != "" {
		bloodType = strings.ReplaceAll(bloodType, " ", "+")
	}

	filter := Domain.EmergencyRequestFilter{
		BloodType:    bloodType,
		Status:       ctx.Query("status"),
		UrgencyLevel: ctx.Query("urgency_level"),
		StartDate:    ctx.Query("start_date"),
		EndDate:      ctx.Query("end_date"),
	}

	if (filter.StartDate != "" && filter.EndDate == "") || (filter.StartDate == "" && filter.EndDate != "") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Both start_date and end_date are required"})
		return
	}

	res, err := c.Usecase.GetAllEmergencies(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *EmergencyRequestController) GetPublishedEmergencies(ctx *gin.Context) {
	reqs, err := c.Usecase.GetPublishedEmergencies()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, reqs)
}

func (c *EmergencyRequestController) GetEmergenciesForDonor(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})
		return
	}

	reqs, err := c.Usecase.GetEmergenciesForDonor(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, reqs)
}
