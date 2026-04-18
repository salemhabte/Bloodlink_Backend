package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"bloodlink/Usecase"
)
//campaign analytics

type CampaignAnalyticsController struct {
	usecase *Usecase.CampaignAnalyticsUsecase
}

func NewCampaignAnalyticsController(u *Usecase.CampaignAnalyticsUsecase) *CampaignAnalyticsController {
	return &CampaignAnalyticsController{usecase: u}
}
func (c *CampaignAnalyticsController) GetDashboard(ctx *gin.Context) {

	stats, err := c.usecase.GetDashboardStats()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
func (c *CampaignAnalyticsController) GetCampaignReport(ctx *gin.Context) {

	id := ctx.Param("id")

	report, err := c.usecase.GetCampaignReport(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, report)
}
func (c *CampaignAnalyticsController) GetAllReports(ctx *gin.Context) {

	data, err := c.usecase.GetAllCampaignReports()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
// blood collector analytics

type CollectorAnalyticsController struct {
	usecase *Usecase.CollectorAnalyticsUsecase
}

func NewCollectorAnalyticsController(u *Usecase.CollectorAnalyticsUsecase) *CollectorAnalyticsController {
	return &CollectorAnalyticsController{usecase: u}
}
func (c *CollectorAnalyticsController) GetKPI(ctx *gin.Context) {

	collectorID := ctx.Query("collector_id")

	data, err := c.usecase.GetCollectorKPI(collectorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
func (c *CollectorAnalyticsController) GetTodayStats(ctx *gin.Context) {

	collectorID := ctx.Query("collector_id")

	if collectorID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "collector_id is required",
		})
		return
	}

	data, err := c.usecase.GetTodayStats(collectorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
func (c *CollectorAnalyticsController) GetDonorInsights(ctx *gin.Context) {

	collectorID := ctx.Query("collector_id")

	data, err := c.usecase.GetDonorInsights(collectorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

//labtech analytics
type LabAnalyticsController struct {
	usecase *Usecase.LabAnalyticsUsecase
}

func NewLabAnalyticsController(u *Usecase.LabAnalyticsUsecase) *LabAnalyticsController {
	return &LabAnalyticsController{usecase: u}
}
func (c *LabAnalyticsController) GetDashboard(ctx *gin.Context) {

	labID := ctx.Query("lab_id")

	data, err := c.usecase.GetDashboard(labID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
