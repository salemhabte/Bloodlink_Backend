package controller

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EmergencyRequestController struct {
	Usecase Interfaces.IEmergencyRequestUsecase
}

func NewEmergencyRequestController(u Interfaces.IEmergencyRequestUsecase) *EmergencyRequestController {
	return &EmergencyRequestController{Usecase: u}
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

	ctx.JSON(http.StatusCreated, gin.H{"message": "Manual emergency request created and published"})
}

func (c *EmergencyRequestController) PublishEmergency(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.Usecase.PublishEmergency(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Emergency request published successfully"})
}

func (c *EmergencyRequestController) RejectEmergency(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.Usecase.RejectEmergency(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Emergency request rejected successfully"})
}

func (c *EmergencyRequestController) GetAllEmergencies(ctx *gin.Context) {
	filter := Domain.EmergencyRequestFilter{
		BloodType:    ctx.Query("blood_type"),
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
