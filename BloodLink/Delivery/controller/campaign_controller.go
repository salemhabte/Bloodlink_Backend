package controller

import (
    "bloodlink/Domain"
    "bloodlink/Usecase"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

type CampaignController struct {
    Usecase *Usecase.CampaignUsecase
}

func NewCampaignController(usecase *Usecase.CampaignUsecase) *CampaignController {
    return &CampaignController{Usecase: usecase}
}

func (c *CampaignController) CreateCampaign(ctx *gin.Context) {
    var input struct {
        Title     string    `json:"title" binding:"required"`
        Content   string    `json:"content" binding:"required"`
        Location  string    `json:"location" binding:"required"`
        StartDate time.Time `json:"start_date" binding:"required"`
        EndDate   time.Time `json:"end_date" binding:"required"`
    }

    if err := ctx.ShouldBindJSON(&input); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    campaign := &Domain.Campaign{
        Title:     input.Title,
        Content:   input.Content,
        Location:  input.Location,
        StartDate: input.StartDate,
        EndDate:   input.EndDate,
    }

    if err := c.Usecase.CreateCampaign(campaign); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusCreated, campaign)
}

func (c *CampaignController) GetAllCampaigns(ctx *gin.Context) {
    campaigns, err := c.Usecase.GetAllCampaigns()
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, campaigns)
}
func (c *CampaignController) GetCampaignByID(ctx *gin.Context) {
	id := ctx.Param("id")

	campaign, err := c.Usecase.GetCampaignByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, campaign)
}

func (c *CampaignController) UpdateCampaign(ctx *gin.Context) {
	id := ctx.Param("id")

	var input struct {
		Title     string    `json:"title"`
		Content   string    `json:"content"`
		Location  string    `json:"location"`
		StartDate time.Time `json:"start_date"`
		EndDate   time.Time `json:"end_date"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	campaign := &Domain.Campaign{
		CampaignID: id,
		Title:      input.Title,
		Content:    input.Content,
		Location:   input.Location,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
	}

	if err := c.Usecase.UpdateCampaign(campaign); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Campaign updated successfully"})
}

func (c *CampaignController) DeleteCampaign(ctx *gin.Context) {
	id := ctx.Param("id")

	if err := c.Usecase.DeleteCampaign(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Campaign deleted successfully"})
}
// Search campaigns by location
func (c *CampaignController) GetCampaignsByLocation(ctx *gin.Context) {

	location := ctx.Query("location")

	campaigns, err := c.Usecase.GetCampaignsByLocation(location)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, campaigns)
}