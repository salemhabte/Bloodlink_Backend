package controller

import (
	Interfaces "bloodlink/Domain/Interfaces"
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationController struct {
	Usecase Interfaces.INotificationUsecase
}

func NewNotificationController(u Interfaces.INotificationUsecase) *NotificationController {
	return &NotificationController{Usecase: u}
}

func (c *NotificationController) GetMyNotifications(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notifs, err := c.Usecase.GetMyNotifications(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, notifs)
}

func (c *NotificationController) MarkAsRead(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	notificationID := ctx.Param("id")
	if err := c.Usecase.MarkAsRead(notificationID, userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func (c *NotificationController) MarkAllAsRead(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := c.Usecase.MarkAllAsRead(userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}
