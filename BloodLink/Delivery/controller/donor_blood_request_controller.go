package controller

import (
	"bloodlink/Usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DonorBloodRequestController struct {
	usecase     *Usecase.DonorBloodRequestUsecase
	userUsecase *Usecase.UserUseCaseBase
}

func NewDonorBloodRequestController(u *Usecase.DonorBloodRequestUsecase, userUsecase *Usecase.UserUseCaseBase) *DonorBloodRequestController {
	return &DonorBloodRequestController{
		usecase:     u,
		userUsecase: userUsecase,
	}
}
func (c *DonorBloodRequestController) CreateRequest(ctx *gin.Context) {
	var req struct {
		DonorID    string `json:"donor_id"`
		QuantityML int    `json:"quantity_ml"`
		Reason     string `json:"reason"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err := c.usecase.CreateRequest(req.DonorID, req.QuantityML, req.Reason)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(201, gin.H{"message": "Request created"})
}
func (c *DonorBloodRequestController) ApproveRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.usecase.ApproveRequest(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Request approved"})
}
func (c *DonorBloodRequestController) FulfillRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.usecase.FulfillRequest(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Request fulfilled"})
}
func (c *DonorBloodRequestController) GetAllRequests(ctx *gin.Context) {
	data, err := c.usecase.GetAllRequests()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(data) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "No blood requests found",
			"data":    []interface{}{},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Requests fetched successfully",
		"data":    data,
	})
}
func (c *DonorBloodRequestController) RejectRequest(ctx *gin.Context) {
	id := ctx.Param("id")

	err := c.usecase.RejectRequest(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to reject request",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Request rejected successfully",
	})
}
func (c *DonorBloodRequestController) GetMyRequests(ctx *gin.Context) {

	userID := ctx.GetString("userID")

	// get donorID from user
	donorID, err := c.userUsecase.GetDonorIDByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "Donor not found",
		})
		return
	}

	data, err := c.usecase.GetMyRequests(donorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch requests",
		})
		return
	}

	if len(data) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "You have no blood requests yet",
			"data":    []interface{}{},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Requests retrieved successfully",
		"data":    data,
	})
}