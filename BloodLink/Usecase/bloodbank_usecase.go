package Usecase

import (
	"bloodlink/Domain"
	Interface "bloodlink/Domain/Interfaces"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ===== CAMPAIGN USECASE =====

type CampaignUsecase struct {
	Repo Interface.ICampaignRepository
}

func NewCampaignUsecase(repo Interface.ICampaignRepository) *CampaignUsecase {
	return &CampaignUsecase{Repo: repo}
}

func (u *CampaignUsecase) CreateCampaign(campaign *Domain.Campaign) error {
	return u.Repo.CreateCampaign(campaign)
}

func (u *CampaignUsecase) GetAllCampaigns() ([]Domain.Campaign, error) {
	return u.Repo.GetAllCampaigns()
}

func (u *CampaignUsecase) GetCampaignByID(id string) (*Domain.Campaign, error) {
	return u.Repo.GetCampaignByID(id)
}

func (u *CampaignUsecase) UpdateCampaign(campaign *Domain.Campaign) error {
	return u.Repo.UpdateCampaign(campaign)
}

func (u *CampaignUsecase) DeleteCampaign(id string) error {
	return u.Repo.DeleteCampaign(id)
}

// Donor Feature
func (u *CampaignUsecase) GetCampaignsByLocation(location string) ([]Domain.Campaign, error) {
	return u.Repo.GetCampaignsByLocation(location)
}


// DonationUsecase contains business logic for blood donations
type DonationUsecase struct {
	repo Interface.IDonationRepository
}

// Constructor
func NewDonationUsecase(repo Interface.IDonationRepository) *DonationUsecase {
	return &DonationUsecase{repo: repo}
}

// CreateDonation handles the business logic for recording a new donation
func (u *DonationUsecase) CreateDonation(record *Domain.DonationRecord) error {

	// 1. Generate donation ID automatically
	record.DonationID = uuid.New().String()

	// 2. Clear client-provided status (system sets it)
	record.Status = ""

	// 3. Set collection date if not provided
	if record.CollectionDate.IsZero() {
		record.CollectionDate = time.Now()
	}

	// 4. Check if donor donated within last 3 months
	lastDonation, err := u.repo.GetLastDonationByDonor(record.DonorID)
	if err == nil && lastDonation != nil {
		if time.Since(lastDonation.CollectionDate).Hours() < 2160 { // 90 days
			return errors.New("donor must wait 3 months before donating again")
		}
	}

	// 5. System automatically evaluates donation status
	u.evaluateDonation(record)

	// 6. Save to database
	return u.repo.CreateDonation(record)
}

// evaluateDonation determines status automatically
func (u *DonationUsecase) evaluateDonation(record *Domain.DonationRecord) {

	if record.Weight < 50 {
		record.Status = "REJECTED_TEMPORARY"
		return
	}

	if record.Hemoglobin < 12 || record.Temperature > 37.5 {
		record.Status = "REJECTED_TEMPORARY"
		return
	}

	record.Status = "APPROVED"
}
// Search donor by email or phone
func (u *DonationUsecase) SearchDonor(query string) (*Domain.DonorResponse, error) {

	if query == "" {
		return nil, errors.New("search value is empty")
	}

	return u.repo.SearchDonor(query)
}

// Update donation status manually by blood collector
func (u *DonationUsecase) UpdateDonationStatus(donationID string, status string) error {
	return u.repo.UpdateDonationStatus(donationID, status)
}

// NEW: Get donation by ID
func (u *DonationUsecase) GetDonationByID(id string) (*Domain.DonationRecord, error) {
	return u.repo.GetDonationByID(id)
}

// NEW: Get all donations
func (u *DonationUsecase) GetAllDonations() ([]Domain.DonationRecord, error) {
	return u.repo.GetAllDonations()
}

// NEW: Update donation medical information
func (u *DonationUsecase) UpdateDonation(record *Domain.DonationRecord) error {

	// Recalculate status after update
	u.evaluateDonation(record)

	return u.repo.UpdateDonation(record)
}
