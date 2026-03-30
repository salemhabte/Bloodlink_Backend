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
	repo Interface.ILabRepository
}

func NewLabUsecase(repo Interface.ILabRepository) *LabUsecase {
	return &LabUsecase{repo: repo}
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
			ExpirationDate: donation.CollectionDate.AddDate(0, 0, 42),
			Status:         "AVAILABLE",
			CreatedAt:      time.Now(),
		}

		if err := u.repo.CreateBloodUnit(bloodUnit); err != nil {
			return err
		}
	}

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
func (u *LabUsecase) UpdateTestResult(result *Domain.DonorTestResult) error {

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

	
	result.DonorID = existing.DonorID

	// 2. Validate (NO override)
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

	// 5. Update donor OverallStatus
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
			fmt.Println("Warning: blood unit not found for donation", result.DonationID)
		}
	} else {
		if bloodUnit != nil {
			if err := u.repo.DeleteBloodUnit(result.DonationID); err != nil {
				return err
			}
		}
	}

	return nil
}
func (u *LabUsecase) RejectBlood(donationID string) error {
	result, err := u.repo.GetTestResult(donationID)
	if err != nil {
		return err
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