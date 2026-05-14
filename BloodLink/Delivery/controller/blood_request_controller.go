package controller

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"bloodlink/Usecase"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type BloodRequestController struct {
	Usecase      Interfaces.IBloodRequestUsecase
	AuditUsecase *Usecase.AuditLogUsecase
}

func NewBloodRequestController(u Interfaces.IBloodRequestUsecase, au *Usecase.AuditLogUsecase) *BloodRequestController {
	return &BloodRequestController{Usecase: u, AuditUsecase: au}
}

func (c *BloodRequestController) CreateBloodRequest(ctx *gin.Context) {
	hospitalAdminID := ctx.GetString("userID")

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
	hospitalAdminID := ctx.GetString("userID")

	bloodType := ctx.Query("blood_type")
	if bloodType != "" {
		bloodType = strings.ReplaceAll(bloodType, " ", "+")
	}

	filter := Domain.BloodRequestFilter{
		HospitalID:   hospitalAdminID,
		BloodType:    bloodType,
		Status:       ctx.Query("status"),
		UrgencyLevel: ctx.Query("urgency_level"),
	}

	reqs, err := c.Usecase.GetHospitalRequests(filter)
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

func (c *BloodRequestController) ApproveRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	oldReq, _ := c.Usecase.GetRequestResponseByID(id)
	oldStatus := "N/A"
	targetName := "Hospital Blood Request"
	if oldReq != nil {
		oldStatus = oldReq.Status
		targetName = fmt.Sprintf("%d units of %s for %s", oldReq.Quantity, oldReq.BloodType, oldReq.HospitalName)
	}

	result, err := c.Usecase.ApproveRequest(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), userID, "APPROVE_HOSPITAL_BLOOD_REQUEST", "HOSPITAL_BLOOD_REQUEST", id, targetName, oldStatus, result.Status)
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *BloodRequestController) RejectRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	oldReq, _ := c.Usecase.GetRequestResponseByID(id)
	oldStatus := "N/A"
	targetName := "Hospital Blood Request"
	if oldReq != nil {
		oldStatus = oldReq.Status
		targetName = fmt.Sprintf("%d units of %s for %s", oldReq.Quantity, oldReq.BloodType, oldReq.HospitalName)
	}

	if err := c.Usecase.RejectRequest(id); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), userID, "REJECT_HOSPITAL_BLOOD_REQUEST", "HOSPITAL_BLOOD_REQUEST", id, targetName, oldStatus, "REJECTED")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Blood request rejected successfully"})
}
