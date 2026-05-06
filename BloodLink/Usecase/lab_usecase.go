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
	result.HepatitisResult = strings.ToUpper(result.HepatitisResult)
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

	// Check for suggestion/conflict
	suggested, conflict := SuggestOverallStatus(result.HIVResult, result.HepatitisResult, result.SyphilisResult, result.OverallStatus)
	if conflict {
		// Return warning to frontend, do not override status automatically
		return fmt.Errorf("⚠ Suggestion: Based on test results, overall status should be '%s'", suggested)
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

	// If CLEARED, create blood unit
	if result.OverallStatus == "CLEARED" {
		bloodUnit := &Domain.BloodUnit{
			BloodUnitID:    uuid.New().String(),
			DonationID:     donation.DonationID,
			BloodType:      result.BloodType,
			VolumeML:       donation.QuantityML,
			CollectionDate: donation.CollectionDate,
			ExpirationDate: calculateExpiration(donation.CollectionDate, result.ComponentType),
ComponentType: result.ComponentType,
			Status:         "AVAILABLE",
			CreatedAt:      time.Now(),
		}

		if err := u.repo.CreateBloodUnit(bloodUnit); err != nil {
			return err
		}
		_ = u.badgeUsecase.EvaluateBadges(donation.DonorID)
	}

	// Notify Donor
	go u.notifUC.SendToDonor(donation.DonorID, "TEST_RESULT", "Test Result Available", "Your blood donation test results are now available. Status: " + result.OverallStatus)

	return nil
}
func SuggestOverallStatus(hiv, hep, syphilis, entered string) (string, bool) {
	// Suggestion based on test results
	var suggested string
	if hiv == "POSITIVE" {
		suggested = "PERMANENTLY_DEFERRED"
	} else if hep == "POSITIVE" || syphilis == "POSITIVE" {
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
	// You need a method in the repository like DeleteBloodUnit(donationID string)
	return u.repo.DeleteBloodUnit(donationID)
}
func (u *LabUsecase) GetTestResult(donationID string) (*Domain.DonorTestResult, error) {
	return u.repo.GetTestResult(donationID)
}

func (u *LabUsecase) GetPendingDonations() ([]Domain.DonationRecord, error) {
	return u.repo.GetPendingDonations()
}
func (u *LabUsecase) GetAllTestResults() ([]Domain.DonorTestResult, error) {
	return u.repo.GetAllTestResults()
}
func (u *LabUsecase) GetTestResultsByStatus(status string) ([]Domain.DonorTestResult, error) {
	return u.repo.GetTestResultsByStatus(status)
}
func (u *LabUsecase) UpdateTestResult(result *Domain.DonorTestResult, currentLabTechID string) error {

	// Normalize input
	result.HIVResult = strings.ToUpper(result.HIVResult)
	result.HepatitisResult = strings.ToUpper(result.HepatitisResult)
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
	// 2. Validate overall status (no override)
	suggested, conflict := SuggestOverallStatus(
		result.HIVResult,
		result.HepatitisResult,
		result.SyphilisResult,
		result.OverallStatus,
	)
	if conflict {
		return fmt.Errorf("invalid overall_status. suggested: %s", suggested)
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

	// 6. Handle blood unit
	bloodUnit, err := u.repo.GetBloodUnitByDonationID(result.DonationID)
	if err != nil {
		bloodUnit = nil
	}

	if result.OverallStatus == "CLEARED" {
		if bloodUnit != nil {
			// Update existing blood unit
			bloodUnit.BloodType = result.BloodType
			bloodUnit.Status = "AVAILABLE"
			if err := u.repo.UpdateBloodUnit(bloodUnit); err != nil {
				return err
			}
		} else {
			// Create new blood unit if it does not exist
			donation, err := u.repo.GetDonationByID(result.DonationID)
			if err != nil {
				return err
			}

			newUnit := &Domain.BloodUnit{
				BloodUnitID:    uuid.New().String(),
				DonationID:     donation.DonationID,
				BloodType:      result.BloodType,
				VolumeML:       donation.QuantityML,
				CollectionDate: donation.CollectionDate,
				ExpirationDate: calculateExpiration(donation.CollectionDate, result.ComponentType),
				Status:         "AVAILABLE",
				CreatedAt:      time.Now(),
			}

			if err := u.repo.CreateBloodUnit(newUnit); err != nil {
				return err
			}
		}
	} else {
		// Remove blood unit if status is TEMPORARY or PERMANENTLY DEFERRED
		if bloodUnit != nil {
			if err := u.repo.DeleteBloodUnit(result.DonationID); err != nil {
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

	// 🔐 SECURITY CHECK
	if result.TestedBy != currentLabTechID {
		return errors.New("you are not allowed to reject another lab tech's test")
	}

	result.OverallStatus = "PERMANENTLY_DEFERRED"

	if err := u.repo.UpdateTestResult(result); err != nil {
		return err
	}

	return u.repo.UpdateDonorOverallStatus(result.DonorID, "PERMANENTLY_DEFERRED")
}
func (u *LabUsecase) GetDonation(donationID string) (*Domain.DonationRecord, error) {
	return u.repo.GetDonationByID(donationID)
}
func calculateExpiration(collectionDate time.Time, component string) time.Time {
	switch component {
	case "WHOLE_BLOOD":
		return collectionDate.AddDate(0, 0, 42)
	case "PLATELETS":
		return collectionDate.AddDate(0, 0, 5)
	case "PLASMA":
		return collectionDate.AddDate(1, 0, 0)
	default:
		return collectionDate.AddDate(0, 0, 35) // safer fallback
	}
}
func (u *LabUsecase) GetMyTestResults(labTechID string) ([]Domain.DonorTestResult, error) {
	return u.repo.GetTestResultsByLabTech(labTechID)
}
func (u *LabUsecase) FilterTestResults(overallStatus, bloodType, componentType string) ([]Domain.DonorTestResult, error) {
	return u.repo.FilterTestResults(overallStatus, bloodType, componentType)
}
func (u *LabUsecase) GetMyTestResultsFiltered(
	labTechID, overallStatus, bloodType, componentType string,
) ([]Domain.DonorTestResult, error) {

	return u.repo.GetMyTestResultsFiltered(
		labTechID,
		overallStatus,
		bloodType,
		componentType,
	)
}
func (u *LabUsecase) GetAllTestsFiltered(
	overallStatus, bloodType, componentType string,
) ([]Domain.DonorTestResult, error) {

	return u.repo.FilterTestResults(
		overallStatus,
		bloodType,
		componentType,
	)
}

func (u *LabUsecase) GetLatestTestResultByDonor(donorID string) (*Domain.DonorTestResult, error) {
	return u.repo.GetLatestTestResultByDonor(donorID)
}