package router

import (
	"bloodlink/Delivery/controller"

	"github.com/gin-gonic/gin"
)

func SetupRouter(campaignController *controller.CampaignController) *gin.Engine {

	r := gin.Default()

	
	// Campaign Routes (Accessible by blood bank admin & Donor)

	campaigns := r.Group("/api/campaigns")
	{
		campaigns.GET("/", campaignController.GetAllCampaigns)
		campaigns.GET("/:id", campaignController.GetCampaignByID)
		campaigns.GET("/search", campaignController.GetCampaignsByLocation)
	}
	 //Campaign Routes Accessible by blood bank Admin

	admin := r.Group("/api/bloodbankadmin")
	{
		adminCampaigns := admin.Group("/campaigns")
		{
			adminCampaigns.POST("/", campaignController.CreateCampaign)
			adminCampaigns.PUT("/:id", campaignController.UpdateCampaign)
			adminCampaigns.DELETE("/:id", campaignController.DeleteCampaign)
		}
	}

	return r
}