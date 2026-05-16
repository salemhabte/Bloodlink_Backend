package controller

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type BloodRequestController struct {
	Usecase Interfaces.IBloodRequestUsecase
}

func NewBloodRequestController(u Interfaces.IBloodRequestUsecase) *BloodRequestController {
	return &BloodRequestController{Usecase: u}
}

func (c *BloodRequestController) CreateBloodRequest(ctx *gin.Context) {
	hospitalAdminID := ctx.GetString("userID")

	var req Domain.CreateBloodRequestBatchDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Usecase.CreateBloodRequest(&req, hospitalAdminID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Blood request submitted successfully"})
}

func (c *BloodRequestController) GetHospitalRequests(ctx *gin.Context) {
	hospitalAdminID := ctx.GetString("userID")

	bloodType := ctx.Query("blood_type")
	if bloodType != "" {
		bloodType = strings.ReplaceAll(bloodType, " ", "+")
	}

	filter := Domain.BloodRequestFilter{
		HospitalID:   hospitalAdminID,
		BloodType:    bloodType,
		Component:    ctx.Query("component"),
		Status:       ctx.Query("status"),
		UrgencyLevel: ctx.Query("urgency_level"),
		StartDate:    ctx.Query("start_date"),
		EndDate:      ctx.Query("end_date"),
	}

	if (filter.StartDate != "" && filter.EndDate == "") || (filter.StartDate == "" && filter.EndDate != "") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Both start_date and end_date are required"})
		return
	}

	res, err := c.Usecase.GetHospitalRequests(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *BloodRequestController) GetAllRequests(ctx *gin.Context) {
	filter := Domain.BloodRequestFilter{
		HospitalID:   ctx.Query("hospital_id"),
		BloodType:    ctx.Query("blood_type"),
		Component:    ctx.Query("component"),
		Status:       ctx.Query("status"),
		UrgencyLevel: ctx.Query("urgency_level"),
		StartDate:    ctx.Query("start_date"),
		EndDate:      ctx.Query("end_date"),
	}

	if (filter.StartDate != "" && filter.EndDate == "") || (filter.StartDate == "" && filter.EndDate != "") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Both start_date and end_date are required"})
		return
	}

	res, err := c.Usecase.GetAllRequests(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *BloodRequestController) ApproveRequest(ctx *gin.Context) {
	requestID := ctx.Param("id")

	result, err := c.Usecase.ApproveRequest(requestID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *BloodRequestController) RejectRequest(ctx *gin.Context) {
	requestID := ctx.Param("id")

	if err := c.Usecase.RejectRequest(requestID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Blood request rejected successfully"})
}
