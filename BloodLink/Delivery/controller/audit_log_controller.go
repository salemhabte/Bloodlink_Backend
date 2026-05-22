package controller

import (
	"bloodlink/Domain"
	"bloodlink/Usecase"
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuditLogController struct {
	usecase *Usecase.AuditLogUsecase
}

func NewAuditLogController(u *Usecase.AuditLogUsecase) *AuditLogController {
	return &AuditLogController{usecase: u}
}

func (c *AuditLogController) GetLogs(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.Query("page"))
	limit, _ := strconv.Atoi(ctx.Query("limit"))

	// Apply defaults so the response reflects the actual values used
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	filter := Domain.AuditLogFilter{
		Action:     ctx.Query("action"),
		TargetType: ctx.Query("target_type"),
		StartDate:  ctx.Query("start_date"),
		EndDate:    ctx.Query("end_date"),
		Page:       page,
		Limit:      limit,
	}

	response, err := c.usecase.GetLogs(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *AuditLogController) GetLogByID(ctx *gin.Context) {
	id := ctx.Param("id")
	log, err := c.usecase.GetLogByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, log)
}

func (c *AuditLogController) DeleteLog(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.usecase.DeleteLog(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Audit log deleted successfully"})
}

func (c *AuditLogController) ExportLogs(ctx *gin.Context) {
	filter := Domain.AuditLogFilter{
		Action:     ctx.Query("action"),
		TargetType: ctx.Query("target_type"),
		StartDate:  ctx.Query("start_date"),
		EndDate:    ctx.Query("end_date"),
	}

	logs, err := c.usecase.ExportLogs(filter)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set headers for CSV file download
	ctx.Header("Content-Disposition", "attachment; filename=audit_logs.csv")
	ctx.Header("Content-Type", "text/csv")
	ctx.Header("Transfer-Encoding", "chunked")

	writer := csv.NewWriter(ctx.Writer)
	defer writer.Flush()

	// Write CSV Header
	writer.Write([]string{"Log ID", "Admin ID", "Action", "Target Type", "Target ID", "Details", "Created At"})

	// Write CSV Data
	for _, log := range logs {
		writer.Write([]string{
			log.LogID,
			log.AdminID,
			log.Action,
			log.TargetType,
			log.TargetID,
			log.Details,
			log.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}
