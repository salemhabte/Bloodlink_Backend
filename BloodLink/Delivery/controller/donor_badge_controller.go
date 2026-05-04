package controller

import (
	"bloodlink/Usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DonorBadgeController struct {
	usecase     *Usecase.DonorBadgeUsecase
	userUsecase *Usecase.UserUseCaseBase
}

func NewDonorBadgeController(
	usecase *Usecase.DonorBadgeUsecase,
	userUsecase *Usecase.UserUseCaseBase,
) *DonorBadgeController {
	return &DonorBadgeController{
		usecase:     usecase,
		userUsecase: userUsecase,
	}
}

// ================= BADGES =================

func (c *DonorBadgeController) GetMyBadges(ctx *gin.Context) {
	userID := ctx.GetString("userID")

	donorID, err := c.userUsecase.GetDonorIDByUserID(ctx, userID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "Donor profile not found",
		})
		return
	}

	badges, err := c.usecase.GetBadges(donorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve badges",
			"error":   err.Error(),
		})
		return
	}

	if len(badges) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "You have no badges yet. Start donating blood to earn rewards 🩸",
			"badges":  []interface{}{},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Badges retrieved successfully",
		"badges":  badges,
	})
}

func (c *DonorBadgeController) GetAllBadges(ctx *gin.Context) {
	badges, err := c.usecase.GetAllBadges()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve badges",
			"error":   err.Error(),
		})
		return
	}

	if len(badges) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "No badges found in the system",
			"badges":  []interface{}{},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "All badges retrieved successfully",
		"badges":  badges,
	})
}

// ================= LEADERBOARD =================

func (c *DonorBadgeController) GetLeaderboard(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	data, err := c.usecase.GetLeaderboard(limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to retrieve leaderboard",
			"error":   err.Error(),
		})
		return
	}

	if len(data) == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"message":     "No donors on the leaderboard yet",
			"leaderboard": []interface{}{},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":     "Leaderboard retrieved successfully",
		"leaderboard": data,
	})
}