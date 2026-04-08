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
		if time.Since(lastDonation.CollectionDate).Hours() < 2160 {
			return errors.New("donor must wait 3 months before donating again")
		}
	}

	// 5. System automatically evaluates donation status
	u.evaluateDonation(record)

	// 6. Save to database
	if err := u.repo.CreateDonation(record); err != nil {
		return err
	}

	// 7.Update donor weight
	if err := u.repo.UpdateDonorWeight(record.DonorID, record.Weight); err != nil {
		return err
	}

	return nil
}
func (u *DonationUsecase) GetPendingDonors() ([]Domain.DonorResponse, error) {
	return u.repo.GetPendingDonors()
}

func (u *DonationUsecase) GetPendingDonorByID(id string) (*Domain.DonorResponse, error) {
	return u.repo.GetPendingDonorByID(id)
}
func (u *DonationUsecase) SearchPendingDonor(query string) (*Domain.DonorResponse, error) {
	return u.repo.SearchPendingDonor(query)
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

    // Get existing donation
    existing, err := u.repo.GetDonationByID(record.DonationID)
    if err != nil {
        return errors.New("donation not found")
    }

    // Prevent wrong donor update
    if existing.DonorID != record.DonorID {
        return errors.New("donor_id does not match this donation")
    }

    // 1. Recalculate donation status
    u.evaluateDonation(record)

    // 2. Update donation
    if err := u.repo.UpdateDonation(record); err != nil {
        return err
    }

    // 3. Update donor weight
    if err := u.repo.UpdateDonorWeight(record.DonorID, record.Weight); err != nil {
        return err
    }

    return nil
}
//BloodInventoryUsecase

type BloodInventoryUsecase struct {
	repo Interface.IBloodInventoryRepository
}

func NewBloodInventoryUsecase(r Interface.IBloodInventoryRepository) *BloodInventoryUsecase {
	return &BloodInventoryUsecase{repo: r}
}

// 🔹 Get All
func (u *BloodInventoryUsecase) GetAllUnits() ([]Domain.BloodUnit, error) {
	return u.repo.GetAllBloodUnits()
}

// 🔹 Get Stats
func (u *BloodInventoryUsecase) GetStats() (map[string]int, error) {
	units, err := u.repo.GetAllBloodUnits()
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"total":      0,
		"available":  0,
		"nearExpiry": 0,
		"expired":    0,
	}

	now := time.Now()

	for _, unit := range units {
		stats["total"]++

		if unit.Status == "AVAILABLE" {
			stats["available"]++
		}

		if unit.ExpirationDate.Before(now) {
			stats["expired"]++
		}

		if unit.ExpirationDate.After(now) &&
			unit.ExpirationDate.Before(now.AddDate(0, 0, 7)) {
			stats["nearExpiry"]++
		}
	}

	return stats, nil
}

// 🔹 Update Status
func (u *BloodInventoryUsecase) UpdateStatus(id, status string) error {
	return u.repo.UpdateBloodUnitStatus(id, status)
}

// 🔹 Delete
func (u *BloodInventoryUsecase) DeleteUnit(id string) error {
	return u.repo.DeleteBloodUnitByID(id)
}
func (u *BloodInventoryUsecase) GetFullDetails(id string) (map[string]interface{}, error) {

	data, err := u.repo.GetFullBloodUnitDetails(id)
	if err != nil {
		return nil, err
	}

	bu := data["blood_unit"].(Domain.BloodUnit)

	now := time.Now()
	diff := bu.ExpirationDate.Sub(now).Hours() / 24

	expiry := map[string]interface{}{
		"days_remaining": int(diff),
		"expires_on":     bu.ExpirationDate,
	}

	//  AUTO STATUS UPDATE
	if bu.ExpirationDate.Before(now) && bu.Status != "EXPIRED" {
		u.repo.UpdateBloodUnitStatus(bu.BloodUnitID, "EXPIRED")
		expiry["expiry_status"] = "EXPIRED"
	} else {
		expiry["expiry_status"] = "VALID"
	}

	data["expiry"] = expiry

	return data, nil
}
func (u *BloodInventoryUsecase) FilterUnits(
	unitID, bloodType, status, startDate, endDate string,
) ([]Domain.BloodUnit, error) {

	return u.repo.FilterBloodUnits(unitID, bloodType, status, startDate, endDate)
}