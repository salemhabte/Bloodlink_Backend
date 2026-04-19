package controller

import (
	"net/http"
	

	"github.com/gin-gonic/gin"
	"bloodlink/Usecase"
)

// ================= ADMIN ANALYTICS CONTROLLER =================

type AdminAnalyticsController struct {
	usecase *Usecase.AdminAnalyticsUsecase
}

func NewAdminAnalyticsController(u *Usecase.AdminAnalyticsUsecase) *AdminAnalyticsController {
	return &AdminAnalyticsController{usecase: u}
}
func (c *AdminAnalyticsController) GetDashboard(ctx *gin.Context) {

	data, err := c.usecase.GetDashboard()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
func (c *AdminAnalyticsController) GetDonorSummary(ctx *gin.Context) {

	data, err := c.usecase.GetDonorSummary()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
func (c *AdminAnalyticsController) GetScreeningSummary(ctx *gin.Context) {

	data, err := c.usecase.GetScreeningSummary()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
func (c *AdminAnalyticsController) GetCollectorSummary(ctx *gin.Context) {

	data, err := c.usecase.GetCollectorSummary()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
func (c *AdminAnalyticsController) GetLabSummary(ctx *gin.Context) {

	data, err := c.usecase.GetLabSummary()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
func (c *AdminAnalyticsController) GetInventorySummary(ctx *gin.Context) {

	data, err := c.usecase.GetInventorySummary()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}