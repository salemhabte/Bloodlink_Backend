package controller

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"bloodlink/Usecase"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type HospitalController struct {
	Usecase Interfaces.IHospitalUsecase
	AuditUsecase *Usecase.AuditLogUsecase
}

func NewHospitalController(u Interfaces.IHospitalUsecase, au *Usecase.AuditLogUsecase) *HospitalController {
	return &HospitalController{Usecase: u, AuditUsecase: au}
}

func (c *HospitalController) SubmitRegistrationRequest(ctx *gin.Context) {
	var req Domain.RegisterHospitalRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Usecase.SubmitRegistrationRequest(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Registration request submitted successfully"})
}

func (c *HospitalController) GetPendingRequests(ctx *gin.Context) {
	reqs, err := c.Usecase.GetPendingRequests()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, reqs)
}

func (c *HospitalController) ApproveRequest(ctx *gin.Context) {
	id := ctx.Param("id")
	adminID := ctx.GetString("userID")

	var req Domain.ApproveHospitalRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldReq, _, _ := c.Usecase.GetHospitalRequestByID(id)
	hospitalName := "Hospital"
	if oldReq != nil {
		hospitalName = oldReq.HospitalName
	}

	if err := c.Usecase.ApproveRequest(id, adminID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if adminID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), adminID, "APPROVE_HOSPITAL_REQUEST", "HOSPITAL_REQUEST", id, hospitalName, "PENDING", "APPROVED")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Hospital request approved and contract drafted"})
}

func (c *HospitalController) RejectRequest(ctx *gin.Context) {
	id := ctx.Param("id")
	oldReq, _, _ := c.Usecase.GetHospitalRequestByID(id)
	hospitalName := "Hospital"
	if oldReq != nil {
		hospitalName = oldReq.HospitalName
	}

	if err := c.Usecase.RejectRequest(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), userID, "REJECT_HOSPITAL_REQUEST", "HOSPITAL_REQUEST", id, hospitalName, "PENDING", "REJECTED")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Hospital request rejected"})
}

func (c *HospitalController) HospitalSignContract(ctx *gin.Context) {
	contractID := ctx.Param("id")
	adminID := ctx.GetString("userID")

	var req Domain.SignContractRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Usecase.HospitalSignContract(contractID, &req, adminID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Contract signed by hospital"})
}

func (c *HospitalController) AdminSignContract(ctx *gin.Context) {
	id := ctx.Param("id")
	adminID := ctx.GetString("userID")

	var req Domain.SignContractRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Usecase.AdminSignContract(id, &req, adminID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	contract, _ := c.Usecase.GetContractByID(id)
	targetName := "Contract"
	if contract != nil {
		targetName = "Contract for Hospital ID " + contract.HospitalID
	}

	if adminID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), adminID, "SIGN_CONTRACT", "CONTRACT", id, targetName, "APPROVED_BY_HOSPITAL", "FINALIZED")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Contract finalized"})
}

func (c *HospitalController) RejectContract(ctx *gin.Context) {
	id := ctx.Param("id")
	userID := ctx.GetString("userID")
	role := ctx.GetString("role")

	contract, _ := c.Usecase.GetContractByID(id)
	oldStatus := "N/A"
	targetName := "Contract"
	if contract != nil {
		oldStatus = contract.Status
		targetName = "Contract for Hospital ID " + contract.HospitalID
	}

	if err := c.Usecase.RejectContract(id, userID, role); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if userID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), userID, "REJECT_CONTRACT", "CONTRACT", id, targetName, oldStatus, "REJECTED")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Contract rejected"})
}

func (c *HospitalController) GetContractByID(ctx *gin.Context) {
	contractID := ctx.Param("id")
	contract, err := c.Usecase.GetContractByID(contractID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Contract not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, contract)
}

func (c *HospitalController) GetMyContracts(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	contracts, err := c.Usecase.GetHospitalContracts(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, contracts)
}

func (c *HospitalController) GetMyLatestContract(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	contract, err := c.Usecase.GetLatestHospitalContract(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "No contracts found for this hospital"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, contract)
}

func (c *HospitalController) DownloadContract(ctx *gin.Context) {
	contractID := ctx.Param("id")
	contract, err := c.Usecase.GetContractByID(contractID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if contract.Document == nil || *contract.Document == "" {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Contract document not found"})
		return
	}

	filePath := *contract.Document
	ctx.FileAttachment(filePath, "contract.pdf")
}

func (c *HospitalController) CreateContractTemplate(ctx *gin.Context) {
	adminID := ctx.GetString("userID")
	var req Domain.CreateTemplateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Usecase.CreateContractTemplate(&req, adminID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if adminID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), adminID, "CREATE_CONTRACT_TEMPLATE", "CONTRACT_TEMPLATE", "NEW", req.Name, "N/A", "CREATED")
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Contract template created"})
}

func (c *HospitalController) GetContractTemplates(ctx *gin.Context) {
	templates, err := c.Usecase.GetContractTemplates()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, templates)
}

func (c *HospitalController) UpdateContractTemplate(ctx *gin.Context) {
	id := ctx.Param("id")
	var req Domain.CreateTemplateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldTemplate, _ := c.Usecase.GetContractTemplateByID(id)
	oldName := "N/A"
	if oldTemplate != nil {
		oldName = oldTemplate.Name
	}

	if err := c.Usecase.UpdateContractTemplate(id, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), userID, "UPDATE_CONTRACT_TEMPLATE", "CONTRACT_TEMPLATE", id, req.Name, oldName, req.Name)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Contract template updated"})
}

func (c *HospitalController) DeleteContractTemplate(ctx *gin.Context) {
	id := ctx.Param("id")

	oldTemplate, _ := c.Usecase.GetContractTemplateByID(id)
	oldName := "N/A"
	if oldTemplate != nil {
		oldName = oldTemplate.Name
	}

	if err := c.Usecase.DeleteContractTemplate(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	userID := ctx.GetString("userID")
	if userID != "" {
		c.AuditUsecase.Log(ctx.Request.Context(), userID, "DELETE_CONTRACT_TEMPLATE", "CONTRACT_TEMPLATE", id, oldName, "EXISTING", "DELETED")
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Contract template deleted"})
}

func (c *HospitalController) GetSignedContracts(ctx *gin.Context) {
	status := strings.ToUpper(strings.TrimSpace(ctx.Query("status")))
	fmt.Printf("GetSignedContracts - Received status: '%s'\n", status)
	contracts, err := c.Usecase.GetSignedContracts(status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, contracts)
}

func (c *HospitalController) GetHospitalDashboard(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	dashboard, err := c.Usecase.GetHospitalDashboard(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, dashboard)
}
