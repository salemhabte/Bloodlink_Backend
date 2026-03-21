package controller

import (
	"context"
	"net/http"

	domain "bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"

	"github.com/gin-gonic/gin"
)

type HospitalController struct {
	HospitalUsecase domainInterface.IHospitalUsecase
}

func NewHospitalController(usecase domainInterface.IHospitalUsecase) *HospitalController {
	return &HospitalController{
		HospitalUsecase: usecase,
	}
}

func (c *HospitalController) RegisterHospital(ctx *gin.Context) {
	var req domain.RegisterHospitalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	hospital, err := c.HospitalUsecase.RegisterHospital(cCtx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":  "Hospital registered successfully",
		"hospital": hospital,
	})
}

func (c *HospitalController) UpdateHospital(ctx *gin.Context) {
	hospitalID := ctx.Param("id")
	if hospitalID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "hospital id is required"})
		return
	}

	var req domain.UpdateHospitalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	hospital, err := c.HospitalUsecase.UpdateHospital(cCtx, hospitalID, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":  "Hospital updated successfully",
		"hospital": hospital,
	})
}

func (c *HospitalController) UploadDocuments(ctx *gin.Context) {
	hospitalID := ctx.Param("id")
	if hospitalID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "hospital id is required"})
		return
	}

	var req domain.UploadHospitalDocumentsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cCtx, cancel := context.WithCancel(ctx.Request.Context())
	defer cancel()

	if err := c.HospitalUsecase.UploadHospitalDocuments(cCtx, hospitalID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Hospital documents uploaded successfully",
	})
}
