package controller

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BloodRequestController struct {
	Usecase Interfaces.IBloodRequestUsecase
}

func NewBloodRequestController(u Interfaces.IBloodRequestUsecase) *BloodRequestController {
	return &BloodRequestController{Usecase: u}
}

func (c *BloodRequestController) CreateBloodRequest(ctx *gin.Context) {
	hospitalAdminID := ctx.GetString("user_id")

	var req Domain.CreateBloodRequestDTO
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
	hospitalAdminID := ctx.GetString("user_id")

	reqs, err := c.Usecase.GetHospitalRequests(hospitalAdminID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, reqs)
}

func (c *BloodRequestController) GetAllRequests(ctx *gin.Context) {
	reqs, err := c.Usecase.GetAllRequests()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, reqs)
}

func (c *BloodRequestController) UpdateStatus(ctx *gin.Context) {
	requestID := ctx.Param("id")

	var req Domain.UpdateBloodRequestStatusDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Usecase.UpdateStatus(requestID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Blood request status updated successfully"})
}
