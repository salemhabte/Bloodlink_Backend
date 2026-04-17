package Usecase

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type IHospitalUserRepository interface {
	CreateUser(ctx context.Context, user *Domain.User) error
}

type hospitalUsecase struct {
	repo       Interfaces.IHospitalRepository
	pdfService IPDFGeneratorService
	userRepo   IHospitalUserRepository
}

func NewHospitalUsecase(repo Interfaces.IHospitalRepository, pdfService IPDFGeneratorService, userRepo IHospitalUserRepository) Interfaces.IHospitalUsecase {
	return &hospitalUsecase{
		repo:       repo,
		pdfService: pdfService,
		userRepo:   userRepo,
	}
}

func (u *hospitalUsecase) SubmitRegistrationRequest(req *Domain.RegisterHospitalRequestDTO) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	requestID := uuid.New().String()
	hospitalReq := &Domain.HospitalRequest{
		RequestID:    requestID,
		HospitalName: req.HospitalName,
		Address:      req.Address,
		Phone:        req.Phone,
		Status:       Domain.RequestStatusPending,
		CreatedAt:    time.Now(),
	}

	err = u.repo.CreateHospitalRequest(hospitalReq)
	if err != nil {
		return err
	}

	adminReq := &Domain.HospitalRequestAdmin{
		RequestAdminID:    uuid.New().String(),
		RequestID:         requestID,
		AdminFullName:     req.AdminFullName,
		AdminEmail:        req.AdminEmail,
		AdminPhone:        req.AdminPhone,
		AdminPasswordHash: string(hashedPassword),
		CreatedAt:         time.Now(),
	}

	return u.repo.CreateHospitalRequestAdmin(adminReq)
}

func (u *hospitalUsecase) GetPendingRequests() ([]Domain.HospitalRequestResponse, error) {
	return u.repo.GetPendingRequests()
}

func (u *hospitalUsecase) ApproveRequest(requestID string, bloodBankAdminID string, payload *Domain.ApproveHospitalRequestDTO) error {
	req, adminReq, err := u.repo.GetHospitalRequestByID(requestID)
	if err != nil {
		return err
	}

	if req.Status != Domain.RequestStatusPending {
		return errors.New("request is not pending")
	}

	// 1. Create real Hospital
	hospitalID := uuid.New().String()
	hospital := &Domain.Hospital{
		HospitalID: hospitalID,
		Name:       req.HospitalName,
		Address:    req.Address,
		Phone:      req.Phone,
		CreatedAt:  time.Now(),
	}
	if err := u.repo.CreateHospital(hospital); err != nil {
		return err
	}

	// 2. Create actual User
	userID := uuid.New().String()
	user := &Domain.User{
		ID:        userID,
		FullName:  adminReq.AdminFullName,
		Email:     adminReq.AdminEmail,
		Phone:     adminReq.AdminPhone,
		Password:  adminReq.AdminPasswordHash,
		Role:      Domain.RoleHospitalAdmin,
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	if err := u.userRepo.CreateUser(context.Background(), user); err != nil {
		return err
	}

	// 3. Create Hospital Admin record
	hospitalAdmin := &Domain.HospitalAdmin{
		HospitalAdminID: uuid.New().String(),
		UserID:          userID,
		HospitalID:      hospitalID,
		CreatedAt:       time.Now(),
	}
	if err := u.repo.CreateHospitalAdmin(hospitalAdmin); err != nil {
		return err
	}

	// 4. Fetch Template and Generate Draft Contract
	template, err := u.repo.GetContractTemplateByID(payload.TemplateID)
	if err != nil {
		return errors.New("contract template not found")
	}

	contractID := uuid.New().String()
	now := time.Now()
	oneYearLater := now.AddDate(1, 0, 0)
	
	renderedText := strings.ReplaceAll(template.Content, "{{hospital_name}}", req.HospitalName)
	renderedText = strings.ReplaceAll(renderedText, "{{contract_start_date}}", now.Format("2006-01-02"))
	renderedText = strings.ReplaceAll(renderedText, "{{contract_end_date}}", oneYearLater.Format("2006-01-02"))

	pdfPath, err := u.pdfService.GenerateDraftContract(contractID, renderedText)
	if err != nil {
		return err
	}

	// 5. Create Contract Record
	contract := &Domain.HospitalContract{
		ContractID:       contractID,
		HospitalID:       hospitalID,
		BloodBankAdminID: bloodBankAdminID,
		Document:         &pdfPath,
		Status:           Domain.ContractStatusPending,
		ContractStart:    &now,
		ContractEnd:      &oneYearLater,
		CreatedAt:        time.Now(),
		TemplateID:       &payload.TemplateID,
	}
	if err := u.repo.CreateContract(contract); err != nil {
		return err
	}

	// 6. Mark Request as Approved
	return u.repo.UpdateHospitalRequestStatus(requestID, Domain.RequestStatusApproved)
}

func (u *hospitalUsecase) RejectRequest(requestID string) error {
	return u.repo.UpdateHospitalRequestStatus(requestID, Domain.RequestStatusRejected)
}

// decodeBase64ToImage saves base64 string to a file and returns the path
func decodeBase64ToImage(base64Str, prefix, id string) (string, error) {
	// Simple stripping of prefix if data:image/... metadata is included
	b64data := base64Str
	if strings.Contains(base64Str, ",") {
		parts := strings.SplitN(base64Str, ",", 2)
		b64data = parts[1]
	}

	data, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		return "", err
	}

	uploadsDir := "uploads/signatures"
	os.MkdirAll(uploadsDir, 0755)

	filePath := filepath.Join(uploadsDir, fmt.Sprintf("%s_%s.png", prefix, id))
	err = os.WriteFile(filePath, data, 0644)
	return filePath, err
}

func (u *hospitalUsecase) HospitalSignContract(contractID string, req *Domain.SignContractRequestDTO, hospitalAdminID string) error {
	contract, err := u.repo.GetContractByID(contractID)
	if err != nil {
		return err
	}

	if contract.Status != Domain.ContractStatusPending {
		return errors.New("contract is not in pending state")
	}

	// Save signature image
	sigPath, err := decodeBase64ToImage(req.SignatureBase64, "hospital", contractID)
	if err != nil {
		return err
	}

	contract.HospitalSignaturePath = &sigPath
	contract.Status = Domain.ContractStatusApprovedByHospital

	return u.repo.UpdateContract(contract)
}

func (u *hospitalUsecase) AdminSignContract(contractID string, req *Domain.SignContractRequestDTO, bloodBankAdminID string) error {
	contract, err := u.repo.GetContractByID(contractID)
	if err != nil {
		return err
	}

	if contract.Status != Domain.ContractStatusApprovedByHospital {
		return errors.New("contract has not been approved by hospital yet")
	}

	// Save signature image
	sigPath, err := decodeBase64ToImage(req.SignatureBase64, "admin", contractID)
	if err != nil {
		return err
	}

	contract.AdminSignaturePath = &sigPath

	// Get Hospital Name
	hospital, err := u.repo.GetHospitalByID(contract.HospitalID)
	if err != nil {
		return err
	}

	// Fetch template to rerender text exactly as it was
	var renderedText string
	if contract.TemplateID != nil {
		template, err := u.repo.GetContractTemplateByID(*contract.TemplateID)
		if err == nil {
			renderedText = strings.ReplaceAll(template.Content, "{{hospital_name}}", hospital.Name)
			renderedText = strings.ReplaceAll(renderedText, "{{contract_start_date}}", contract.ContractStart.Format("2006-01-02"))
			renderedText = strings.ReplaceAll(renderedText, "{{contract_end_date}}", contract.ContractEnd.Format("2006-01-02"))
		}
	}
	if renderedText == "" {
		renderedText = fmt.Sprintf("This blood supply contract is made and entered into on %s between the centralized Blood Bank and %s.", contract.ContractStart.Format("2006-01-02"), hospital.Name)
	}

	// Regenerate PDF with both signatures
	finalPdfPath, err := u.pdfService.GenerateFinalContract(
		contractID,
		renderedText,
		*contract.HospitalSignaturePath,
		*contract.AdminSignaturePath,
	)
	if err != nil {
		return err
	}

	contract.Status = Domain.ContractStatusFinalized
	contract.Document = &finalPdfPath

	return u.repo.UpdateContract(contract)
}

func (u *hospitalUsecase) RejectContract(contractID string, userID string, role string) error {
	contract, err := u.repo.GetContractByID(contractID)
	if err != nil {
		return err
	}

	if contract.Status == Domain.ContractStatusFinalized {
		return errors.New("cannot reject a finalized contract")
	}

	contract.Status = Domain.ContractStatusRejected
	return u.repo.UpdateContract(contract)
}

func (u *hospitalUsecase) CreateContractTemplate(req *Domain.CreateTemplateRequestDTO, adminID string) error {
	t := &Domain.ContractTemplate{
		TemplateID: uuid.New().String(),
		Name:       req.Name,
		Content:    req.Content,
		CreatedBy:  &adminID,
		CreatedAt:  time.Now(),
	}
	return u.repo.CreateContractTemplate(t)
}

func (u *hospitalUsecase) GetContractTemplates() ([]Domain.ContractTemplate, error) {
	return u.repo.GetContractTemplates()
}

func (u *hospitalUsecase) UpdateContractTemplate(templateID string, req *Domain.CreateTemplateRequestDTO) error {
	t, err := u.repo.GetContractTemplateByID(templateID)
	if err != nil {
		return err
	}
	t.Name = req.Name
	t.Content = req.Content
	return u.repo.UpdateContractTemplate(t)
}

func (u *hospitalUsecase) DeleteContractTemplate(templateID string) error {
	return u.repo.DeleteContractTemplate(templateID)
}
