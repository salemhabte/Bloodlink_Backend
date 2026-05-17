package Usecase

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"os"
	"regexp"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type IHospitalUserRepository interface {
	CreateUser(ctx context.Context, user *Domain.User) error
	GetUserByPhone(ctx context.Context, phone string) (*Domain.User, error)
	GetDonorByUserID(ctx context.Context, userID string) (*Domain.Donor, error)
}

type hospitalUsecase struct {
	repo         Interfaces.IHospitalRepository
	pdfService   IPDFGeneratorService
	userRepo     IHospitalUserRepository
	notifUC      Interfaces.INotificationUsecase
	donationRepo Interfaces.IDonationRepository
}

func NewHospitalUsecase(repo Interfaces.IHospitalRepository, pdfService IPDFGeneratorService, userRepo IHospitalUserRepository, notifUC Interfaces.INotificationUsecase, donationRepo Interfaces.IDonationRepository) Interfaces.IHospitalUsecase {
	return &hospitalUsecase{
		repo:         repo,
		pdfService:   pdfService,
		userRepo:     userRepo,
		notifUC:      notifUC,
		donationRepo: donationRepo,
	}
}

func (u *hospitalUsecase) SubmitRegistrationRequest(req *Domain.RegisterHospitalRequestDTO) error {
	// Validate phone formats first
	phoneRegex := `^\+251[79]\d{8}$`
	re := regexp.MustCompile(phoneRegex)
	if !re.MatchString(req.Phone) {
		return errors.New("hospital phone number must follow the +251 x[7,9]xxxxxxx format (e.g. +251912345678)")
	}
	if !re.MatchString(req.AdminPhone) {
		return errors.New("admin phone number must follow the +251 x[7,9]xxxxxxx format (e.g. +251912345678)")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Check if hospital phone already exists in hospitals table
	existingHospital, _ := u.repo.GetHospitalByPhone(req.Phone)
	if existingHospital != nil {
		return errors.New("a hospital with this phone number is already registered")
	}

	// Check if hospital phone already exists in users table (across all roles)
	existingUser, _ := u.userRepo.GetUserByPhone(context.Background(), req.Phone)
	if existingUser != nil {
		return errors.New("this phone number is already associated with another account")
	}

	// Check if admin phone already exists
	existingAdmin, _ := u.userRepo.GetUserByPhone(context.Background(), req.AdminPhone)
	if existingAdmin != nil {
		return errors.New("admin phone number already registered")
	}

	requestID := uuid.New().String()
	hospitalReq := &Domain.HospitalRequest{
		RequestID:       requestID,
		HospitalName:    req.HospitalName,
		Address:         req.Address,
		Phone:           req.Phone,
		LicenseDocument: req.LicenseDocument,
		Status:          Domain.RequestStatusPending,
		CreatedAt:       time.Now(),
		Latitude:        req.Latitude,
		Longitude:       req.Longitude,
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

	err = u.repo.CreateHospitalRequestAdmin(adminReq)
	if err == nil {
		go u.notifUC.SendToRole(Domain.RoleBloodBankAdmin, "CONTRACT", "New Hospital Registration", fmt.Sprintf("A new registration request was submitted by %s", req.HospitalName))
	}
	return err
}

func (u *hospitalUsecase) GetPendingRequests(filter Domain.HospitalRequestFilter) (*Domain.HospitalRequestListResponse, error) {
	requests, err := u.repo.GetPendingRequests(filter)
	if err != nil {
		return nil, err
	}

	var analytics Domain.SummaryAnalytics
	analytics.TotalRequests = len(requests)
	for _, r := range requests {
		switch r.Status {
		case Domain.RequestStatusApproved:
			analytics.TotalFulfilled++
		case Domain.RequestStatusPending:
			analytics.TotalPending++
		case Domain.RequestStatusRejected:
			analytics.TotalCancelled++
		}
	}

	return &Domain.HospitalRequestListResponse{
		Requests:  requests,
		Analytics: analytics,
	}, nil
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
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
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

func (u *hospitalUsecase) HospitalSignContract(contractID string, req *Domain.SignContractRequestDTO, hospitalAdminID string) error {
	contract, err := u.repo.GetContractByID(contractID)
	if err != nil {
		return err
	}

	if contract.Status != Domain.ContractStatusPending {
		return errors.New("contract is not in pending state")
	}

	contract.HospitalSignaturePath = &req.SignatureURL
	contract.Status = Domain.ContractStatusApprovedByHospital

	err = u.repo.UpdateContract(contract)
	if err == nil {
		go u.notifUC.SendToRole(Domain.RoleBloodBankAdmin, "CONTRACT", "Contract Signed by Hospital", "A hospital has signed their contract and is waiting for your signature.")
	}
	return err
}

func (u *hospitalUsecase) AdminSignContract(contractID string, req *Domain.SignContractRequestDTO, bloodBankAdminID string) error {
	contract, err := u.repo.GetContractByID(contractID)
	if err != nil {
		return err
	}

	if contract.Status != Domain.ContractStatusApprovedByHospital {
		return errors.New("contract has not been approved by hospital yet")
	}

	contract.AdminSignaturePath = &req.SignatureURL

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
	finalPdfPath, err := u.generateFinalPDF(contract, hospital)
	if err != nil {
		return err
	}

	contract.Status = Domain.ContractStatusFinalized
	contract.Document = &finalPdfPath

	err = u.repo.UpdateContract(contract)
	if err == nil {
		go u.notifUC.SendToHospital(contract.HospitalID, "CONTRACT", "Contract Finalized", "Your hospital contract has been finalized by the Blood Bank Admin.")
	}
	return err
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
	err = u.repo.UpdateContract(contract)
	if err == nil {
		go u.notifUC.SendToHospital(contract.HospitalID, "CONTRACT", "Contract Rejected", "Your hospital contract has been rejected.")
	}
	return err
}

func (u *hospitalUsecase) GetContractByID(contractID string) (*Domain.HospitalContract, error) {
	contract, err := u.repo.GetContractByID(contractID)
	if err != nil {
		return nil, err
	}

	// Check if document exists on disk. If not, re-generate it.
	// This is important for ephemeral storage like Render.
	if contract.Document != nil && *contract.Document != "" {
		if _, err := os.Stat(*contract.Document); os.IsNotExist(err) {
			switch contract.Status {
			case Domain.ContractStatusFinalized:
				hospital, _ := u.repo.GetHospitalByID(contract.HospitalID)
				if hospital != nil {
					newPath, err := u.generateFinalPDF(contract, hospital)
					if err == nil {
						contract.Document = &newPath
						_ = u.repo.UpdateContract(contract)
					}
				}
			case Domain.ContractStatusPending, Domain.ContractStatusApprovedByHospital:
				// Re-render draft if possible
				if contract.TemplateID != nil {
					template, err := u.repo.GetContractTemplateByID(*contract.TemplateID)
					if err == nil {
						hospital, _ := u.repo.GetHospitalByID(contract.HospitalID)
						hName := "Hospital"
						if hospital != nil {
							hName = hospital.Name
						}
						renderedText := strings.ReplaceAll(template.Content, "{{hospital_name}}", hName)
						renderedText = strings.ReplaceAll(renderedText, "{{contract_start_date}}", contract.ContractStart.Format("2006-01-02"))
						renderedText = strings.ReplaceAll(renderedText, "{{contract_end_date}}", contract.ContractEnd.Format("2006-01-02"))

						newPath, err := u.pdfService.GenerateDraftContract(contract.ContractID, renderedText)
						if err == nil {
							contract.Document = &newPath
							_ = u.repo.UpdateContract(contract)
						}
					}
				}
			}
		}
	}

	return contract, nil
}

func (u *hospitalUsecase) generateFinalPDF(contract *Domain.HospitalContract, hospital *Domain.Hospital) (string, error) {
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

	return u.pdfService.GenerateFinalContract(
		contract.ContractID,
		renderedText,
		*contract.HospitalSignaturePath,
		*contract.AdminSignaturePath,
	)
}

func (u *hospitalUsecase) GetHospitalContracts(userID string) (*Domain.HospitalContractListResponse, error) {
	admin, err := u.repo.GetHospitalAdminByUserID(userID)
	if err != nil {
		return nil, err
	}
	contracts, err := u.repo.GetContractsByHospitalID(admin.HospitalID)
	if err != nil {
		return nil, err
	}

	var responses []Domain.HospitalContractResponse
	for _, c := range contracts {
		responses = append(responses, Domain.HospitalContractResponse{
			ContractID:            c.ContractID,
			HospitalID:            c.HospitalID,
			BloodBankAdminID:      c.BloodBankAdminID,
			Document:              c.Document,
			Status:                c.Status,
			ContractStart:         c.ContractStart,
			ContractEnd:           c.ContractEnd,
			CreatedAt:             c.CreatedAt,
			HospitalSignaturePath: c.HospitalSignaturePath,
			AdminSignaturePath:    c.AdminSignaturePath,
		})
	}

	return &Domain.HospitalContractListResponse{
		Contracts: responses,
		Analytics: u.calculateContractAnalytics(responses),
	}, nil
}

func (u *hospitalUsecase) calculateContractAnalytics(contracts []Domain.HospitalContractResponse) Domain.ContractAnalytics {
	var analytics Domain.ContractAnalytics
	analytics.TotalContracts = len(contracts)
	now := time.Now()
	for _, c := range contracts {
		switch c.Status {
		case Domain.ContractStatusFinalized:
			if c.ContractEnd != nil && c.ContractEnd.Before(now) {
				analytics.TotalExpired++
			} else {
				analytics.TotalActive++
			}
		case Domain.ContractStatusPending, Domain.ContractStatusApprovedByHospital:
			analytics.TotalPending++
		case Domain.ContractStatusRejected:
			analytics.TotalRejected++
		}
	}
	return analytics
}
func (u *hospitalUsecase) GetAllHospitals() (*Domain.HospitalListResponse, error) {
	hospitals, err := u.repo.GetAllHospitals()
	if err != nil {
		return nil, err
	}

	var analytics Domain.HospitalAnalyticsSummary
	analytics.TotalHospitals = len(hospitals)
	// We could count active vs inactive if we had a status, 
	// for now let's just assume all in the table are active or count by some other metric.
	// Maybe we can count hospitals with active contracts.
	
	return &Domain.HospitalListResponse{
		Hospitals: hospitals,
		Analytics: analytics,
	}, nil
}
func (u *hospitalUsecase) GetLatestHospitalContract(userID string) (*Domain.HospitalContract, error) {
	admin, err := u.repo.GetHospitalAdminByUserID(userID)
	if err != nil {
		return nil, err
	}
	contracts, err := u.repo.GetContractsByHospitalID(admin.HospitalID)
	if err != nil {
		return nil, err
	}
	if len(contracts) == 0 {
		return nil, sql.ErrNoRows
	}
	return &contracts[0], nil
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

func (u *hospitalUsecase) GetSignedContracts(status string) (*Domain.HospitalContractListResponse, error) {
	contracts, err := u.repo.GetSignedContracts(status)
	if err != nil {
		return nil, err
	}

	return &Domain.HospitalContractListResponse{
		Contracts: contracts,
		Analytics: u.calculateContractAnalytics(contracts),
	}, nil
}

func (u *hospitalUsecase) GetHospitalDashboard(userID string) (*Domain.HospitalDashboard, error) {
	admin, err := u.repo.GetHospitalAdminByUserID(userID)
	if err != nil {
		return nil, err
	}
	return u.repo.GetHospitalDashboard(admin.HospitalID)
}

func (u *hospitalUsecase) ConfirmHospitalDonation(donorPhone string, hospitalAdminUserID string) error {
	_, err := u.repo.GetHospitalAdminByUserID(hospitalAdminUserID)
	if err != nil {
		return errors.New("unauthorized: not a hospital admin")
	}

	user, err := u.userRepo.GetUserByPhone(context.Background(), donorPhone)
	if err != nil || user == nil {
		return errors.New("donor not found with the given phone number")
	}

	donor, err := u.userRepo.GetDonorByUserID(context.Background(), user.ID)
	if err != nil || donor == nil {
		return errors.New("user is not registered as a donor")
	}

	// Record donation
	record := &Domain.DonationRecord{
		DonationID:     uuid.New().String(),
		DonorID:        donor.DonorID,
		CollectedBy:    hospitalAdminUserID,
		CollectionDate: time.Now(),
		QuantityML:     450, // default standard for hospital direct donations
		Status:         "APPROVED",
		OverallStatus:  "CLEARED",
		CreatedAt:      time.Now(),
	}

	err = u.donationRepo.CreateDonation(record)
	if err != nil {
		return err
	}

	// Update overall_status
	_ = u.donationRepo.UpdateDonorOverallStatus(donor.DonorID, "CLEARED")

	return nil
}
