package Usecase

import (
	"bloodlink/Domain"
	Interface "bloodlink/Domain/Interfaces"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LabUsecase struct {
	repo          Interface.ILabRepository
	badgeUsecase  *DonorBadgeUsecase
	notifUC       Interface.INotificationUsecase
}

func NewLabUsecase(
	repo Interface.ILabRepository,
	badgeUsecase *DonorBadgeUsecase,
	notifUC Interface.INotificationUsecase,
) *LabUsecase {
	return &LabUsecase{
		repo:         repo,
		badgeUsecase: badgeUsecase,
		notifUC:      notifUC,
	}
}

func (u *LabUsecase) ProcessTestResult(result *Domain.DonorTestResult) error {
	result.HIVResult = strings.ToUpper(result.HIVResult)
	result.HepatitisBResult = strings.ToUpper(result.HepatitisBResult)
	result.HepatitisCResult = strings.ToUpper(result.HepatitisCResult)
	result.SyphilisResult = strings.ToUpper(result.SyphilisResult)
	result.OverallStatus = strings.ToUpper(result.OverallStatus)
	result.BloodType = strings.ToUpper(result.BloodType)

	// Prevent duplicate test creation
	existing, _ := u.repo.GetTestResult(result.DonationID)
	if existing != nil {
		return errors.New("a test for this donation already exists")
	}

	// Get donation info
	donation, err := u.repo.GetDonationByID(result.DonationID)
	if err != nil {
		return err
	}

	result.TestID = uuid.New().String()
	result.DonorID = donation.DonorID
	result.CreatedAt = time.Now()

	// Check for suggestion/conflict (4-disease check, no TEMPORARILY_DEFERRED)
	suggested, conflict := SuggestOverallStatus(result.HIVResult, result.HepatitisBResult, result.HepatitisCResult, result.SyphilisResult, result.OverallStatus)
	if conflict {
		return fmt.Errorf("⚠ Suggestion: Based on test results, overall status should be '%s'", suggested)
	}

	// === VALIDATION for CLEARED status ===
	if result.OverallStatus == "CLEARED" {
		// Must have components
		if len(result.Components) == 0 {
			return errors.New("components are required when overall_status is CLEARED")
		}

		// Quantity-based component validation
		if donation.QuantityML == 350 {
			if len(result.Components) != 1 {
				return errors.New("350ml donations must have exactly 1 component (WHOLE_BLOOD)")
			}
			compType := strings.ToUpper(result.Components[0].ComponentType)
			if compType != "WHOLE_BLOOD" {
				return errors.New("350ml donations must be processed as WHOLE_BLOOD")
			}
		} else if donation.QuantityML == 450 {
			if len(result.Components) < 1 || len(result.Components) > 3 {
				return errors.New("450ml donations must have between 1 and 3 components")
			}
			validTypes := map[string]bool{"PRBC": true, "CRBC": true, "PLATELETS": true, "PLASMA": true}
			for _, comp := range result.Components {
				ct := strings.ToUpper(comp.ComponentType)
				if !validTypes[ct] {
					return fmt.Errorf("invalid component_type for 450ml: %s (must be PRBC, PLATELETS, or PLASMA)", comp.ComponentType)
				}
				if comp.Quantity <= 0 {
					return fmt.Errorf("component quantity must be greater than 0 for %s", comp.ComponentType)
				}
			}
		}

		// Quantity validation: sum of components <= donation quantity
		totalQuantity := 0
		for _, comp := range result.Components {
			totalQuantity += comp.Quantity
		}
		if totalQuantity > donation.QuantityML {
			return fmt.Errorf("total component quantity (%d mL) exceeds donation quantity (%d mL)", totalQuantity, donation.QuantityML)
		}

		// Storage fields required for CLEARED
		if strings.TrimSpace(result.StorageLocation) == "" {
			return errors.New("storage_location is required when overall_status is CLEARED")
		}

		posMap := make(map[string]bool)
		for _, comp := range result.Components {
			pos := strings.TrimSpace(comp.PositionNumber)
			if pos == "" {
				return errors.New("position_number is required for all components when overall_status is CLEARED")
			}
			if posMap[pos] {
				return fmt.Errorf("duplicate position_number '%s' found in request", pos)
			}
			posMap[pos] = true

			// Check if slot is taken
			occupied, err := u.repo.IsSlotOccupied(result.StorageLocation, result.RackNumber, result.ShelfNumber, pos)
			if err != nil {
				return err
			}
			if occupied {
				return fmt.Errorf("Slot [Rack %s, Shelf %s, Pos %s] in %s is already occupied", result.RackNumber, result.ShelfNumber, pos, result.StorageLocation)
			}
		}

		// Capacity Check (Max 12 slots per cell)
		occupiedCount, err := u.repo.GetOccupiedSlotCount(result.StorageLocation, result.RackNumber, result.ShelfNumber)
		if err != nil {
			return err
		}
		if 12 - occupiedCount < len(result.Components) {
			return fmt.Errorf("Only %d positions available in this cell. You are trying to store %d components.", 12-occupiedCount, len(result.Components))
		}
	}

	// === VALIDATION for PERMANENTLY_DEFERRED ===
	if result.OverallStatus == "PERMANENTLY_DEFERRED" {
		if len(result.Components) > 0 {
			return errors.New("components must be empty when overall_status is PERMANENTLY_DEFERRED")
		}
		if strings.TrimSpace(result.StorageLocation) != "" || strings.TrimSpace(result.RackNumber) != "" || strings.TrimSpace(result.ShelfNumber) != "" {
			return errors.New("storage fields must be empty when overall_status is PERMANENTLY_DEFERRED")
		}
	}

	// === VALIDATION for TEMPORARILY_DEFERRED ===
	if result.OverallStatus == "TEMPORARILY_DEFERRED" {
		if len(result.Components) > 0 {
			return errors.New("components must be empty when overall_status is TEMPORARILY_DEFERRED")
		}
		if strings.TrimSpace(result.StorageLocation) != "" || strings.TrimSpace(result.RackNumber) != "" || strings.TrimSpace(result.ShelfNumber) != "" {
			return errors.New("storage fields must be empty when overall_status is TEMPORARILY_DEFERRED")
		}
	}

	// Save test result
	if err := u.repo.CreateTestResult(result); err != nil {
		return err
	}

	// Update donor blood type
	if err := u.repo.UpdateDonorBloodType(donation.DonorID, result.BloodType); err != nil {
		return err
	}

	// Update donor status
	if err := u.repo.UpdateDonorOverallStatus(donation.DonorID, result.OverallStatus); err != nil {
		return err
	}

	// If CLEARED, create blood units (one per component)
	if result.OverallStatus == "CLEARED" {
		for _, comp := range result.Components {
			componentType := strings.ToUpper(comp.ComponentType)
			// Normalize CRBC → PRBC for storage
			if componentType == "CRBC" {
				componentType = "PRBC"
			}

			bloodUnit := &Domain.BloodUnit{
				BloodUnitID:     uuid.New().String(),
				DonationID:      donation.DonationID,
				BloodType:       result.BloodType,
				ComponentType:   componentType,
				QuantityML:        comp.Quantity,
				CollectionDate:  donation.CollectionDate,
				ExpirationDate:  calculateExpiration(donation.CollectionDate, componentType),
				Status:          "AVAILABLE",
				StorageLocation: result.StorageLocation,
				RackNumber:      result.RackNumber,
				ShelfNumber:     result.ShelfNumber,
				PositionNumber:  comp.PositionNumber,
				CreatedAt:       time.Now(),
			}

			if err := u.repo.CreateBloodUnit(bloodUnit); err != nil {
				return err
			}
		}
		_ = u.badgeUsecase.EvaluateBadges(donation.DonorID)
	}

	// Notify Donor
	go u.notifUC.SendToDonor(donation.DonorID, "TEST_RESULT", "Test Result Available", "Your blood donation test results are now available. Status: "+result.OverallStatus)

	return nil
}

// SuggestOverallStatus checks 4 disease markers and suggests the correct overall_status.
func SuggestOverallStatus(hiv, hepB, hepC, syphilis, entered string) (string, bool) {
	var suggested string
	if hiv == "POSITIVE" || hepB == "POSITIVE" || hepC == "POSITIVE" {
		suggested = "PERMANENTLY_DEFERRED"
	} else if syphilis == "POSITIVE" {
		suggested = "TEMPORARILY_DEFERRED"
	} else {
		suggested = "CLEARED"
	}

	if entered != suggested {
		return suggested, true // true means there's a conflict
	}
	return suggested, false
}

func (u *LabUsecase) removeBloodUnit(donationID string) error {
	return u.repo.DeleteBloodUnit(donationID)
}

func (u *LabUsecase) GetTestResult(donationID string) (*Domain.DonorTestResult, error) {
	return u.repo.GetTestResult(donationID)
}

func (u *LabUsecase) GetPendingDonations() ([]Domain.DonationRecord, error) {
	return u.repo.GetPendingDonations()
}

func (u *LabUsecase) GetTestResultsByStatus(status string) ([]Domain.DonorTestResult, error) {
	return u.repo.GetTestResultsByStatus(status)
}

func (u *LabUsecase) UpdateTestResult(result *Domain.DonorTestResult, currentLabTechID string) error {

	// Normalize input
	result.HIVResult = strings.ToUpper(result.HIVResult)
	result.HepatitisBResult = strings.ToUpper(result.HepatitisBResult)
	result.HepatitisCResult = strings.ToUpper(result.HepatitisCResult)
	result.SyphilisResult = strings.ToUpper(result.SyphilisResult)
	result.OverallStatus = strings.ToUpper(result.OverallStatus)
	result.BloodType = strings.ToUpper(result.BloodType)

	fmt.Println("Updating donation:", result.DonationID)

	// 1. Check if test exists
	existing, err := u.repo.GetTestResult(result.DonationID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("test result not found")
	}

	// SECURITY CHECK — ensure same lab tech owns the test
	if existing.TestedBy != currentLabTechID {
		return errors.New("you are not allowed to edit another lab tech's test result")
	}

	// 2. Validate overall status (no override) — 4 disease check
	suggested, conflict := SuggestOverallStatus(
		result.HIVResult,
		result.HepatitisBResult,
		result.HepatitisCResult,
		result.SyphilisResult,
		result.OverallStatus,
	)
	if conflict {
		return fmt.Errorf("invalid overall_status. suggested: %s", suggested)
	}

	// Get donation info for quantity validation
	donation, err := u.repo.GetDonationByID(result.DonationID)
	if err != nil {
		return err
	}

	// === VALIDATION for CLEARED status ===
	if result.OverallStatus == "CLEARED" {
		if len(result.Components) == 0 {
			return errors.New("components are required when overall_status is CLEARED")
		}

		// Quantity-based component validation
		if donation.QuantityML == 350 {
			if len(result.Components) != 1 {
				return errors.New("350ml donations must have exactly 1 component (WHOLE_BLOOD)")
			}
			compType := strings.ToUpper(result.Components[0].ComponentType)
			if compType != "WHOLE_BLOOD" {
				return errors.New("350ml donations must be processed as WHOLE_BLOOD")
			}
		} else if donation.QuantityML == 450 {
			if len(result.Components) < 1 || len(result.Components) > 3 {
				return errors.New("450ml donations must have between 1 and 3 components")
			}
			validTypes := map[string]bool{"PRBC": true, "CRBC": true, "PLATELETS": true, "PLASMA": true}
			for _, comp := range result.Components {
				ct := strings.ToUpper(comp.ComponentType)
				if !validTypes[ct] {
					return fmt.Errorf("invalid component_type for 450ml: %s (must be PRBC, PLATELETS, or PLASMA)", comp.ComponentType)
				}
				if comp.Quantity <= 0 {
					return fmt.Errorf("component quantity must be greater than 0 for %s", comp.ComponentType)
				}
			}
		}

		totalQuantity := 0
		for _, comp := range result.Components {
			totalQuantity += comp.Quantity
		}
		if totalQuantity > donation.QuantityML {
			return fmt.Errorf("total component quantity (%d mL) exceeds donation quantity (%d mL)", totalQuantity, donation.QuantityML)
		}
		if strings.TrimSpace(result.StorageLocation) == "" {
			return errors.New("storage_location is required when overall_status is CLEARED")
		}

		posMap := make(map[string]bool)
		for _, comp := range result.Components {
			pos := strings.TrimSpace(comp.PositionNumber)
			if pos == "" {
				return errors.New("position_number is required for all components when overall_status is CLEARED")
			}
			if posMap[pos] {
				return fmt.Errorf("duplicate position_number '%s' found in request", pos)
			}
			posMap[pos] = true

			// Check if slot is taken
			occupied, err := u.repo.IsSlotOccupied(result.StorageLocation, result.RackNumber, result.ShelfNumber, pos)
			if err != nil {
				return err
			}
			if occupied {
				// If we are updating, it's possible the occupied slot belongs to the SAME donation we are replacing.
				// We need to check if the unit occupying this slot belongs to this donation.
				// If so, it's okay because we will delete it. We can just ignore the occupancy error.
				oldUnits, _ := u.repo.GetBloodUnitsByDonationID(result.DonationID)
				isSelf := false
				for _, ou := range oldUnits {
					if !ou.IsDeleted && ou.StorageLocation == result.StorageLocation && ou.RackNumber == result.RackNumber && ou.ShelfNumber == result.ShelfNumber && ou.PositionNumber == pos {
						isSelf = true
						break
					}
				}
				if !isSelf {
					return fmt.Errorf("Slot [Rack %s, Shelf %s, Pos %s] in %s is already occupied", result.RackNumber, result.ShelfNumber, pos, result.StorageLocation)
				}
			}
		}

		// Capacity Check (Max 12 slots per cell)
		occupiedCount, err := u.repo.GetOccupiedSlotCount(result.StorageLocation, result.RackNumber, result.ShelfNumber)
		if err != nil {
			return err
		}
		
		// If updating, subtract the units from the SAME donation that are in this same cell 
		// because they will be deleted.
		oldUnits, _ := u.repo.GetBloodUnitsByDonationID(result.DonationID)
		for _, ou := range oldUnits {
			if !ou.IsDeleted && ou.Status != "USED" && ou.StorageLocation == result.StorageLocation && ou.RackNumber == result.RackNumber && ou.ShelfNumber == result.ShelfNumber {
				occupiedCount--
			}
		}

		if 12 - occupiedCount < len(result.Components) {
			return fmt.Errorf("Only %d positions available in this cell. You are trying to store %d components.", 12-occupiedCount, len(result.Components))
		}
	}

	if result.OverallStatus == "PERMANENTLY_DEFERRED" {
		if len(result.Components) > 0 {
			return errors.New("components must be empty when overall_status is PERMANENTLY_DEFERRED")
		}
	}

	if result.OverallStatus == "TEMPORARILY_DEFERRED" {
		if len(result.Components) > 0 {
			return errors.New("components must be empty when overall_status is TEMPORARILY_DEFERRED")
		}
	}

	// 3. Update donor test result
	if err := u.repo.UpdateTestResult(result); err != nil {
		return err
	}

	// 4. Update donor blood type
	if err := u.repo.UpdateDonorBloodType(result.DonorID, result.BloodType); err != nil {
		return err
	}

	// 5. Update donor overall status
	if err := u.repo.UpdateDonorOverallStatus(result.DonorID, result.OverallStatus); err != nil {
		return err
	}

	// 6. Handle blood units — delete old ones first, then re-create if CLEARED
	_ = u.repo.DeleteBloodUnitsByDonationID(result.DonationID)

	if result.OverallStatus == "CLEARED" {
		for _, comp := range result.Components {
			componentType := strings.ToUpper(comp.ComponentType)
			if componentType == "CRBC" {
				componentType = "PRBC"
			}

			bloodUnit := &Domain.BloodUnit{
				BloodUnitID:     uuid.New().String(),
				DonationID:      donation.DonationID,
				BloodType:       result.BloodType,
				ComponentType:   componentType,
				QuantityML:        comp.Quantity,
				CollectionDate:  donation.CollectionDate,
				ExpirationDate:  calculateExpiration(donation.CollectionDate, componentType),
				Status:          "AVAILABLE",
				StorageLocation: result.StorageLocation,
				RackNumber:      result.RackNumber,
				ShelfNumber:     result.ShelfNumber,
				PositionNumber:  comp.PositionNumber,
				CreatedAt:       time.Now(),
			}

			if err := u.repo.CreateBloodUnit(bloodUnit); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *LabUsecase) RejectBlood(donationID string, currentLabTechID string) error {

	result, err := u.repo.GetTestResult(donationID)
	if err != nil {
		return err
	}

	// SECURITY CHECK
	if result.TestedBy != currentLabTechID {
		return errors.New("you are not allowed to reject another lab tech's test")
	}

	result.OverallStatus = "PERMANENTLY_DEFERRED"

	if err := u.repo.UpdateTestResult(result); err != nil {
		return err
	}

	// Delete blood units since blood is rejected
	_ = u.repo.DeleteBloodUnitsByDonationID(donationID)

	return u.repo.UpdateDonorOverallStatus(result.DonorID, "PERMANENTLY_DEFERRED")
}

func (u *LabUsecase) GetPendingDonation(donationID string) (*Domain.DonationRecord, error) {
	return u.repo.GetPendingDonationByID(donationID)
}

func calculateExpiration(collectionDate time.Time, component string) time.Time {
	switch component {
	case "PRBC", "CRBC":
		return collectionDate.AddDate(0, 0, 42)
	case "PLATELETS":
		return collectionDate.AddDate(0, 0, 5)
	case "PLASMA":
		return collectionDate.AddDate(1, 0, 0)
	case "CRYOPRECIPITATE":
		return collectionDate.AddDate(1, 0, 0)
	case "WHOLE_BLOOD":
		return collectionDate.AddDate(0, 0, 42)
	default:
		return collectionDate.AddDate(0, 0, 35) // safer fallback
	}
}



func (u *LabUsecase) GetMyTestResultsFiltered(filter Domain.TestFilter) ([]Domain.DonorTestResult, error) {
	tests, err := u.repo.GetMyTestResultsFiltered(filter)
	if err != nil {
		return nil, err
	}

	for i := range tests {
		u.populateTestComponents(&tests[i])
	}
	return tests, nil
}

func (u *LabUsecase) GetAllTestsFiltered(filter Domain.TestFilter) ([]Domain.DonorTestResult, error) {

	tests, err := u.repo.FilterTestResults(filter)
	if err != nil {
		return nil, err
	}

	for i := range tests {
		u.populateTestComponents(&tests[i])
	}
	return tests, nil
}

func (u *LabUsecase) populateTestComponents(test *Domain.DonorTestResult) {
	if test.OverallStatus == "CLEARED" {
		units, err := u.repo.GetBloodUnitsByDonationID(test.DonationID)
		if err == nil && len(units) > 0 {
			var comps []Domain.BloodComponentInput
			for _, unit := range units {
				comps = append(comps, Domain.BloodComponentInput{
					ComponentType:  unit.ComponentType,
					Quantity:       unit.QuantityML,
					PositionNumber: unit.PositionNumber,
				})
			}
			test.Components = comps
			test.StorageLocation = units[0].StorageLocation
			test.RackNumber = units[0].RackNumber
			test.ShelfNumber = units[0].ShelfNumber
		}
	}
}

func (u *LabUsecase) GetLatestTestResultByDonor(donorID string) (*Domain.DonorTestResult, error) {
	return u.repo.GetLatestTestResultByDonor(donorID)
}