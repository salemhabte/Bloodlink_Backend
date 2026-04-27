package controller

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type HospitalController struct {
	Usecase Interfaces.IHospitalUsecase
}

func NewHospitalController(u Interfaces.IHospitalUsecase) *HospitalController {
	return &HospitalController{Usecase: u}
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
	requestID := ctx.Param("id")
	bloodBankAdminID := ctx.GetString("userID")

	var payload Domain.ApproveHospitalRequestDTO
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Usecase.ApproveRequest(requestID, bloodBankAdminID, &payload); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Hospital request approved and contract drafted"})
}

func (c *HospitalController) RejectRequest(ctx *gin.Context) {
	requestID := ctx.Param("id")
	if err := c.Usecase.RejectRequest(requestID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
	contractID := ctx.Param("id")
	adminID := ctx.GetString("userID")

	var req Domain.SignContractRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Usecase.AdminSignContract(contractID, &req, adminID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Contract finalized"})
}

func (c *HospitalController) RejectContract(ctx *gin.Context) {
	contractID := ctx.Param("id")
	userID := ctx.GetString("userID")
	role := ctx.GetString("role")

	if err := c.Usecase.RejectContract(contractID, userID, role); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
	templateID := ctx.Param("id")
	var req Domain.CreateTemplateRequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Usecase.UpdateContractTemplate(templateID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "Contract template updated"})
}

func (c *HospitalController) DeleteContractTemplate(ctx *gin.Context) {
	templateID := ctx.Param("id")
	if err := c.Usecase.DeleteContractTemplate(templateID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
