package Usecase

import (
	"bloodlink/Domain"
	Interface "bloodlink/Domain/Interfaces"
	"errors"
	"fmt"
	"log"
	"strings"
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
	// Validate campaign start date
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if campaign.StartDate.Before(today) {
		return errors.New("campaign start date cannot be in the past")
	}

	err := u.Repo.CreateCampaign(campaign)
	if err == nil {
		go u.NotifUC.SendToRole(Domain.RoleDonor, "CAMPAIGN", "New Blood Drive", "A new campaign '"+campaign.Title+"' has been created at "+campaign.Location)
	}
	return err
}

func (u *CampaignUsecase) GetAllCampaigns(filter Domain.CampaignFilter, liveOnly bool) (*Domain.CampaignListResponse, error) {
	campaigns, err := u.Repo.GetAllCampaigns(filter, liveOnly)
	if err != nil {
		return nil, err
	}

	response := &Domain.CampaignListResponse{
		Campaigns: campaigns,
	}

	now := time.Now()
	nearLimit := now.Add(168 * time.Hour) // 7 days

	for _, c := range campaigns {
		if c.EndDate.After(now) {
			if c.StartDate.After(now) {
				response.UpcomingCampaigns++
			} else {
				response.OngoingCampaigns++
				if c.EndDate.Before(nearLimit) {
					response.ClosingSoonCampaigns++
				}
			}
		} else {
			response.ClosedCampaigns++
		}
	}

	// For Admin, we might want the true total from DB, 
	// but here we can just count the current slice if no filters are applied.
	// For now, let's just use the slice length as 'TotalCampaigns' for Admin if not liveOnly.
	if !liveOnly {
		response.TotalCampaigns = len(campaigns)
	}

	return response, nil
}

func (u *CampaignUsecase) GetCampaignByID(id string) (*Domain.Campaign, error) {
	return u.Repo.GetCampaignByID(id)
}

func (u *CampaignUsecase) GetLiveCampaignByID(id string) (*Domain.Campaign, error) {
	return u.Repo.GetLiveCampaignByID(id)
}

func (u *CampaignUsecase) UpdateCampaign(campaign *Domain.Campaign) error {
	return u.Repo.UpdateCampaign(campaign)
}

func (u *CampaignUsecase) DeleteCampaign(id string) error {
	return u.Repo.DeleteCampaign(id)
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
	// 2. Validate status field is provided
	// ================================
	record.Status = strings.ToUpper(strings.TrimSpace(record.Status))
	if record.Status != "APPROVED" && record.Status != "REJECTED_TEMPORARY" {
		return errors.New("status must be APPROVED or REJECTED_TEMPORARY")
	}
	
	if record.Status == "REJECTED_TEMPORARY" && strings.TrimSpace(record.RejectionReason) == "" {
		return errors.New("rejection_reason is required when status is REJECTED_TEMPORARY")
	}

	if record.Status == "APPROVED" {
		record.RejectionReason = "" // Clear reason if approved
	}

	// ================================
	// 3. Validate donation quantity
	// ================================
	if record.Status == "REJECTED_TEMPORARY" {
		record.QuantityML = 0 // Donors who are rejected do not donate
	} else if record.QuantityML != 350 && record.QuantityML != 450 {
		return errors.New("quantity_ml must be 350 or 450 for approved donations")
	}

	// ================================
	// 4. Set collection date
	// ================================
	if record.CollectionDate.IsZero() {
		return errors.New("collection_date is required")
	} else if record.CollectionDate.After(time.Now().Add(1 * time.Minute)) {
		return errors.New("collection_date cannot be in the future")
	}

	// ================================
	// 5. Validate campaign (if provided)
	// ================================
	if record.CampaignID != nil {
		_, err := u.campaignRepo.GetLiveCampaignByID(*record.CampaignID)
		if err != nil {
			return err // Will return "campaign is already closed" if past
		}
	}

	// ================================
	// 6. IMPORTANT: Check donor eligibility FIRST
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
	// 7. 3-MONTH RULE CHECK
	// ================================
	lastDonation, err := u.repo.GetLastDonationByDonor(record.DonorID)
	if err == nil && lastDonation != nil {
		if time.Since(lastDonation.CollectionDate).Hours() < 2160 {
			return errors.New("donor must wait 3 months before donating again")
		}
	}

	// ================================
	// 8. Suggestion/Review flow for status
	// ================================
	suggested, conflict := SuggestDonationStatus(record.Weight, record.Hemoglobin, record.Temperature, record.Pulse, record.BloodPressure)
	if conflict(record.Status) {
		return fmt.Errorf("⚠ Suggestion: Based on screening values, status should be '%s'", suggested)
	}

	// ================================
	// 9. Save donation
	// ================================
	if err := u.repo.CreateDonation(record); err != nil {
		return err
	}

	// ================================
	// 10. POST-DONATION UPDATES
	// ================================
	// Reset overall_status to Pending (waiting for new lab results)
	_ = u.repo.UpdateDonorOverallStatus(record.DonorID, "Pending")

	// Update donor weight
	if err := u.repo.UpdateDonorWeight(record.DonorID, record.Weight); err != nil {
		log.Printf("[ERROR] Failed to update donor weight: %v", err)
	}

	// Notify Lab Techs only if donation is APPROVED
	if record.Status == "APPROVED" {
		log.Printf("[DEBUG] Triggering notification for lab tech...")
		go u.notifUC.SendToRole(Domain.RoleLabTech, "DONATION", "New Donation", "A new donation record is pending lab testing.")
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

// SuggestDonationStatus evaluates screening data and returns the suggested status
// plus a function to check if a given entered status conflicts with the suggestion.
// WHO thresholds: weight >= 50kg, hemoglobin >= 12.5 g/dL, temp <= 37.5°C, pulse 50-100bpm
func SuggestDonationStatus(weight float64, hemoglobin float64, temperature float64, pulse int, bloodPressure string) (string, func(string) bool) {
	suggested := "APPROVED"

	if weight < 50 {
		suggested = "REJECTED_TEMPORARY"
	} else if hemoglobin < 12.5 {
		suggested = "REJECTED_TEMPORARY"
	} else if temperature > 37.5 {
		suggested = "REJECTED_TEMPORARY"
	} else if pulse < 50 || pulse > 100 {
		suggested = "REJECTED_TEMPORARY"
	}
	// Blood pressure parsing: systolic 90-160, diastolic 60-100
	if bloodPressure != "" {
		parts := strings.Split(bloodPressure, "/")
		if len(parts) == 2 {
			var systolic, diastolic int
			fmt.Sscanf(parts[0], "%d", &systolic)
			fmt.Sscanf(parts[1], "%d", &diastolic)
			if systolic < 90 || systolic > 160 || diastolic < 60 || diastolic > 100 {
				suggested = "REJECTED_TEMPORARY"
			}
		}
	}

	return suggested, func(entered string) bool {
		return entered != suggested
	}
}
// Search donor by email or phone
func (u *DonationUsecase) SearchDonor(query string) (*Domain.DonorResponse, error) {

	if query == "" {
		return nil, errors.New("search value is empty")
	}

	return u.repo.SearchDonor(query)
}

// Update donation status manually by blood collector
func (u *DonationUsecase) UpdateDonationStatus(donationID string, status string, rejectionReason string, collectorID string) error {

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
	// 2. Validate Status
	// ================================
	status = strings.ToUpper(strings.TrimSpace(status))
	validStatuses := map[string]bool{"APPROVED": true, "REJECTED_TEMPORARY": true}
	if !validStatuses[status] {
		return errors.New("invalid status: must be APPROVED or REJECTED_TEMPORARY")
	}
	
	if status == "REJECTED_TEMPORARY" && strings.TrimSpace(rejectionReason) == "" {
		return errors.New("rejection_reason is required when status is REJECTED_TEMPORARY")
	}

	if status == "APPROVED" {
		rejectionReason = ""
	}

	// 3. Suggestion check
	suggested, conflict := SuggestDonationStatus(existing.Weight, existing.Hemoglobin, existing.Temperature, existing.Pulse, existing.BloodPressure)
	if conflict(status) {
		return fmt.Errorf("⚠ Suggestion: Based on screening values, status should be '%s'", suggested)
	}

	// ================================
	// 3. Update status in DB
	// ================================
	existing.Status = status
	existing.RejectionReason = rejectionReason
	return u.repo.UpdateDonation(existing)
}

// NEW: Get donation by ID
func (u *DonationUsecase) GetDonationByID(id string) (*Domain.DonationRecord, error) {
	return u.repo.GetDonationByID(id)
}



// NEW: Get all donations by donor ID
func (u *DonationUsecase) GetAllDonationsByDonor(donorID string) ([]Domain.DonationRecord, error) {
	return u.repo.GetAllDonationsByDonor(donorID)
}

// UpdateDonation updates donation medical information
func (u *DonationUsecase) UpdateDonation(record *Domain.DonationRecord) error {

	// Validate and check status via suggestion flow
	record.Status = strings.ToUpper(strings.TrimSpace(record.Status))
	if record.Status != "APPROVED" && record.Status != "REJECTED_TEMPORARY" {
		return errors.New("status must be APPROVED or REJECTED_TEMPORARY")
	}

	if record.Status == "REJECTED_TEMPORARY" && strings.TrimSpace(record.RejectionReason) == "" {
		return errors.New("rejection_reason is required when status is REJECTED_TEMPORARY")
	}

	if record.Status == "APPROVED" {
		record.RejectionReason = ""
	}

	// Validate quantity
	if record.Status == "REJECTED_TEMPORARY" {
		record.QuantityML = 0
	} else if record.QuantityML != 350 && record.QuantityML != 450 {
		return errors.New("quantity_ml must be 350 or 450 for approved donations")
	}

	if record.CollectionDate.After(time.Now().Add(1 * time.Minute)) {
		return errors.New("collection_date cannot be in the future")
	}

	// Get existing donation
	existing, err := u.repo.GetDonationByID(record.DonationID)
	if err != nil {
		return errors.New("donation not found")
	}

	// Identify donor from the existing record (no need for frontend to pass donor_id)
	record.DonorID = existing.DonorID

	// Only the collector who created this donation can update it
	if existing.CollectedBy != record.CollectedBy {
		return errors.New("you are not allowed to update this donation")
	}

	suggested, conflict := SuggestDonationStatus(record.Weight, record.Hemoglobin, record.Temperature, record.Pulse, record.BloodPressure)
	if conflict(record.Status) {
		return fmt.Errorf("⚠ Suggestion: Based on screening values, status should be '%s'", suggested)
	}

	// Update donation
	if err := u.repo.UpdateDonation(record); err != nil {
		return err
	}

	// Update donor weight
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
func (u *BloodInventoryUsecase) GetAllUnits(filter Domain.BloodUnitFilter) ([]Domain.BloodUnit, error) {
	return u.repo.GetAllBloodUnits(filter)
}

// 🔹 Get Stats
func (u *BloodInventoryUsecase) GetStats() (map[string]int, error) {
	units, err := u.repo.GetAllBloodUnits(Domain.BloodUnitFilter{})
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

func (u *BloodInventoryUsecase) GetByID(id string) (*Domain.BloodUnit, error) {
	return u.repo.GetBloodUnitByID(id)
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
	unit, err := u.repo.GetBloodUnitByID(id)
	if err != nil {
		return err
	}

	if unit.Status != "USED" && unit.Status != "EXPIRED" {
		return errors.New("only USED or EXPIRED blood units can be deleted")
	}

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
func (u *BloodInventoryUsecase) FilterUnits(filter Domain.BloodUnitFilter) ([]Domain.BloodUnit, error) {
	return u.repo.FilterBloodUnits(filter)
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

func (u *BloodInventoryUsecase) ConvertPlasmaToCryo(plasmaUnitID string, cryoQuantity int, cryoPoorQuantity *int) error {
	plasma, err := u.repo.GetBloodUnitByID(plasmaUnitID)
	if err != nil {
		return err
	}

	if plasma.ComponentType != "PLASMA" {
		return errors.New("only PLASMA units can be converted to Cryoprecipitate")
	}
	if plasma.Status != "AVAILABLE" {
		return errors.New("only AVAILABLE units can be converted")
	}
	if cryoQuantity <= 0 || cryoQuantity >= plasma.QuantityML {
		return errors.New("cryoprecipitate quantity must be greater than 0 and less than total plasma quantity")
	}

	// Calculate default poor plasma quantity
	finalCryoPoorQuantity := plasma.QuantityML - cryoQuantity

	// If user provided a specific quantity, use it (handle loss)
	if cryoPoorQuantity != nil {
		if *cryoPoorQuantity > finalCryoPoorQuantity {
			return errors.New("cryo-poor plasma quantity cannot exceed the remaining plasma quantity")
		}
		if *cryoPoorQuantity < 0 {
			return errors.New("cryo-poor plasma quantity cannot be negative")
		}
		finalCryoPoorQuantity = *cryoPoorQuantity
	}

	now := time.Now()
	
	// Create new Cryoprecipitate unit
	cryoUnit := &Domain.BloodUnit{
		BloodUnitID:     uuid.New().String(),
		DonationID:      plasma.DonationID,
		BloodType:       plasma.BloodType,
		ComponentType:   "CRYOPRECIPITATE",
		QuantityML:        cryoQuantity,
		CollectionDate:  plasma.CollectionDate,
		ExpirationDate:  plasma.ExpirationDate, // Or calculate dynamically if different
		Status:          "AVAILABLE",
		StorageLocation: plasma.StorageLocation,
		RackNumber:      plasma.RackNumber,
		ShelfNumber:     plasma.ShelfNumber,
		CreatedAt:       now,
	}

	// Create new Cryo-poor Plasma unit
	cryoPoorUnit := &Domain.BloodUnit{
		BloodUnitID:     uuid.New().String(),
		DonationID:      plasma.DonationID,
		BloodType:       plasma.BloodType,
		ComponentType:   "CRYO_POOR_PLASMA",
		QuantityML:        finalCryoPoorQuantity,
		CollectionDate:  plasma.CollectionDate,
		ExpirationDate:  plasma.ExpirationDate, // Or calculate dynamically if different
		Status:          "AVAILABLE",
		StorageLocation: plasma.StorageLocation,
		RackNumber:      plasma.RackNumber,
		ShelfNumber:     plasma.ShelfNumber,
		CreatedAt:       now,
	}

	return u.repo.ConvertPlasmaToCryo(plasmaUnitID, cryoUnit, cryoPoorUnit)
}