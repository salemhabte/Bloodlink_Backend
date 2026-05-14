package controller

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"bloodlink/Usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EmergencyRequestController struct {
	Usecase Interfaces.IEmergencyRequestUsecase
	AuditUsecase *Usecase.AuditLogUsecase
}

func NewEmergencyRequestController(u Interfaces.IEmergencyRequestUsecase, au *Usecase.AuditLogUsecase) *EmergencyRequestController {
	return &EmergencyRequestController{Usecase: u, AuditUsecase: au}
}

func (c *EmergencyRequestController) CreateManualEmergency(ctx *gin.Context) {
	var dto Domain.CreateEmergencyRequestDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Usecase.CreateManualEmergency(&dto); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), userID, "CREATE_EMERGENCY", "EMERGENCY_REQUEST", "NEW", dto.BloodType+" Emergency for "+dto.HospitalName, "N/A", "PUBLISHED")
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Manual emergency request created and published"})
}

func (c *EmergencyRequestController) PublishEmergency(ctx *gin.Context) {
	id := ctx.Param("id")
	oldEmergency, _ := c.Usecase.GetEmergencyByID(id)
	oldStatus := "N/A"
	targetName := "Emergency Request"
	if oldEmergency != nil {
		oldStatus = oldEmergency.Status
		targetName = oldEmergency.BloodType + " Emergency for " + oldEmergency.HospitalName
	}

	if err := c.Usecase.PublishEmergency(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), userID, "PUBLISH_EMERGENCY", "EMERGENCY_REQUEST", id, targetName, oldStatus, "PUBLISHED")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Emergency request published successfully"})
}

func (c *EmergencyRequestController) RejectEmergency(ctx *gin.Context) {
	id := ctx.Param("id")

	oldEmergency, _ := c.Usecase.GetEmergencyByID(id)
	oldStatus := "N/A"
	targetName := "Emergency Request"
	if oldEmergency != nil {
		oldStatus = oldEmergency.Status
		targetName = oldEmergency.BloodType + " Emergency for " + oldEmergency.HospitalName
	}

	if err := c.Usecase.RejectEmergency(id); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), userID, "REJECT_EMERGENCY", "EMERGENCY_REQUEST", id, targetName, oldStatus, "REJECTED")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Emergency request rejected successfully"})
}

func (c *EmergencyRequestController) GetAllEmergencies(ctx *gin.Context) {
	reqs, err := c.Usecase.GetAllEmergencies()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, reqs)
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
