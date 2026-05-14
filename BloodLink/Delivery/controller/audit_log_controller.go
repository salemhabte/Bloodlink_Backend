package controller

import (
	"bloodlink/Domain"
	"bloodlink/Usecase"
	"net/http"

	"strconv"
	"github.com/gin-gonic/gin"
)

type AuditLogController struct {
	Usecase *Usecase.AuditLogUsecase
}

func NewAuditLogController(u *Usecase.AuditLogUsecase) *AuditLogController {
	return &AuditLogController{Usecase: u}
}

func (c *AuditLogController) GetLogs(ctx *gin.Context) {
	filter := Domain.AuditLogFilter{
		UserID:     ctx.Query("user_id"),
		Action:     ctx.Query("action"),
		TargetType: ctx.Query("target_type"),
		StartDate:  ctx.Query("start_date"),
		EndDate:    ctx.Query("end_date"),
	}

	logs, err := c.Usecase.GetLogs(ctx.Request.Context(), filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(logs) == 0 {
		ctx.JSON(http.StatusOK, gin.H{"message": "No audit logs found for the given criteria"})
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

// GetLogByID returns a single audit log entry by its ID.
func (c *AuditLogController) GetLogByID(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	logEntry, err := c.Usecase.GetLogByID(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if logEntry == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Audit log not found"})
		return
	}
	ctx.JSON(http.StatusOK, logEntry)
}

// DeleteLog removes an audit log entry.
func (c *AuditLogController) DeleteLog(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := c.Usecase.DeleteLog(ctx.Request.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Audit log deleted", "log_id": id})
}
