package router

import (
	"bloodlink/Delivery/controller"

	"github.com/gin-gonic/gin"
)

func SetupRouter(campaignController *controller.CampaignController) *gin.Engine {
	r := gin.Default()

	admin := r.Group("/api/admin")
	{
		campaigns := admin.Group("/campaigns")
		{
			campaigns.POST("/", campaignController.CreateCampaign)
			campaigns.GET("/", campaignController.GetAllCampaigns)
			campaigns.GET("/:id", campaignController.GetCampaignByID)
			campaigns.PUT("/:id", campaignController.UpdateCampaign)
			campaigns.DELETE("/:id", campaignController.DeleteCampaign)
		}
	}

	return r
}