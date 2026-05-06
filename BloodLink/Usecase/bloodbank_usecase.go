package Usecase

import (
	"bloodlink/Domain"
	Interface "bloodlink/Domain/Interfaces"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

// ===== CAMPAIGN USECASE =====

type CampaignUsecase struct {
	Repo    Interface.ICampaignRepository
	NotifUC Interface.INotificationUsecase
}

func NewCampaignUsecase(repo Interface.ICampaignRepository, notifUC Interface.INotificationUsecase) *CampaignUsecase {
	return &CampaignUsecase{Repo: repo, NotifUC: notifUC}
}

func (u *CampaignUsecase) CreateCampaign(campaign *Domain.Campaign) error {
	err := u.Repo.CreateCampaign(campaign)
	if err == nil {
		go u.NotifUC.SendToRole(Domain.RoleDonor, "CAMPAIGN", "New Blood Drive", "A new campaign '"+campaign.Title+"' has been created at "+campaign.Location)
	}
	return err
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
	repo         Interface.IDonationRepository
	campaignRepo Interface.ICampaignRepository
	notifUC      Interface.INotificationUsecase
}

// Constructor
func NewDonationUsecase(repo Interface.IDonationRepository, campaignRepo Interface.ICampaignRepository, notifUC Interface.INotificationUsecase) *DonationUsecase {
	return &DonationUsecase{repo: repo, campaignRepo: campaignRepo, notifUC: notifUC}
}

// CreateDonation handles the business logic for recording a new donation
func (u *DonationUsecase) CreateDonation(record *Domain.DonationRecord) error {

	// ================================
	// 1. Generate donation ID
	// ================================
	record.DonationID = uuid.New().String()

	// ================================
	// 2. System-controlled status
	// ================================
	record.Status = "PENDING"

	// ================================
	// 3. Set collection date
	// ================================
	if record.CollectionDate.IsZero() {
		record.CollectionDate = time.Now()
	}

	// ================================
	// 4. Validate campaign (if provided)
	// ================================
	if record.CampaignID != nil {
		_, err := u.campaignRepo.GetCampaignByID(*record.CampaignID)
		if err != nil {
			return errors.New("invalid campaign_id")
		}
	}

	// ================================
	// 5. IMPORTANT: Check donor eligibility FIRST
	// ================================
	overallStatus, err := u.repo.GetDonorOverallStatus(record.DonorID)
	if err != nil {
		return errors.New("donor not found")
	}

	//  BLOCK permanently deferred donors (e.g HIV positive)
	if overallStatus == "PERMANENTLY_DEFERRED" {
		return errors.New("donor is permanently deferred and cannot donate")
	}

	// ================================
	// 6. 3-MONTH RULE CHECK
	// ================================
	lastDonation, err := u.repo.GetLastDonationByDonor(record.DonorID)
	if err == nil && lastDonation != nil {
		if time.Since(lastDonation.CollectionDate).Hours() < 2160 {
			return errors.New("donor must wait 3 months before donating again")
		}
	}

	// ================================
	// 7. Evaluate donation (medical screening)
	// ================================
	u.evaluateDonation(record)



	// ================================
	// 8. Save donation
	// ================================
	if err := u.repo.CreateDonation(record); err != nil {
		return err
	}
	
	// Notify Lab Techs
	log.Printf("[DEBUG] Triggering notification for lab tech...")
	go u.notifUC.SendToRole(Domain.RoleLabTech, "DONATION", "New Donation", "A new donation record is pending lab testing.")

	// ================================
	// 9. Update donor weight
	// ================================
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
func (u *DonationUsecase) UpdateDonationStatus(donationID string, status string, collectorID string) error {

	// ================================
	// 1. Get existing donation
	// ================================
	existing, err := u.repo.GetDonationByID(donationID)
	if err != nil {
		return errors.New("donation not found")
	}

	// ==========================================================
	//  SECURITY CHECK (IMPORTANT)
	// Only creator collector can update status
	// ==========================================================
	if existing.CollectedBy != collectorID {
		return errors.New("you are not allowed to update this donation status")
	}

	// ================================
	// 2. Update status in DB
	// ================================
	return u.repo.UpdateDonationStatus(donationID, status)
}

// NEW: Get donation by ID
func (u *DonationUsecase) GetDonationByID(id string) (*Domain.DonationRecord, error) {
	return u.repo.GetDonationByID(id)
}



// NEW: Get all donations by donor ID
func (u *DonationUsecase) GetAllDonationsByDonor(donorID string) ([]Domain.DonationRecord, error) {
	return u.repo.GetAllDonationsByDonor(donorID)
}

// NEW: Update donation medical information
func (u *DonationUsecase) UpdateDonation(record *Domain.DonationRecord) error {

    // Get existing donation
    existing, err := u.repo.GetDonationByID(record.DonationID)
    if err != nil {
        return errors.New("donation not found")
    }
	// Only the collector who created this donation can update it
	if existing.CollectedBy != record.CollectedBy {
		return errors.New("you are not allowed to update this donation")
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


func (u *DonationUsecase) GetAllDonations(filter Domain.DonationFilter) ([]Domain.DonationRecord, error) {
	return u.repo.GetDonations(filter)
}
func (u *DonationUsecase) GetMyDonations(collectorID string, filter Domain.DonationFilter) ([]Domain.DonationRecord, error) {
    filter.CollectorID = collectorID
    return u.repo.GetDonations(filter)
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
		"reserved":   0,
		"used":       0,
	}

	now := time.Now()

	for _, unit := range units {
		stats["total"]++

		isRealTimeExpired := unit.ExpirationDate.Before(now) && unit.Status != "EXPIRED" && unit.Status != "USED"

		if isRealTimeExpired || unit.Status == "EXPIRED" {
			stats["expired"]++
		} else if unit.Status == "AVAILABLE" {
			stats["available"]++
			
			if unit.ExpirationDate.After(now) && unit.ExpirationDate.Before(now.AddDate(0, 0, 7)) {
				stats["nearExpiry"]++
			}
		} else if unit.Status == "RESERVED" {
			stats["reserved"]++
		} else if unit.Status == "USED" {
			stats["used"]++
		}
	}

	return stats, nil
}

// 🔹 Update Status
func (u *BloodInventoryUsecase) UpdateStatus(id, status string) error {
	return u.repo.UpdateBloodUnitStatus(id, status)
}

// 🔹 Mark as Used (only if currently reserved)
func (u *BloodInventoryUsecase) MarkUnitAsUsed(id string) error {
	return u.repo.MarkUnitAsUsed(id)
}

// 🔹 Delete with Audit (only if expired or used)
func (u *BloodInventoryUsecase) DeleteUnit(id string) error {
	return u.repo.DeleteWithAudit(id)
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
	if bu.ExpirationDate.Before(now) && bu.Status != "EXPIRED" && bu.Status != "USED" {
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
func (u *BloodInventoryUsecase) UpdateExpiredUnits() error {
	return u.repo.MarkExpiredUnits()
}

func (u *BloodInventoryUsecase) GetReservedUnitsByHospital(hospitalID string) ([]Domain.BloodUnit, error) {
	return u.repo.GetReservedUnitsByHospitalID(hospitalID)
}

func (u *BloodInventoryUsecase) ExpireReservations() ([]string, error) {
	// Cutoff is 24 hours ago
	cutoff := time.Now().Add(-24 * time.Hour)
	return u.repo.ExpireStaleReservations(cutoff)
}