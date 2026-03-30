package Domain

import "bloodlink/Domain"

type ILabRepository interface {
	CreateTestResult(result *Domain.DonorTestResult) error
	CreateBloodUnit(unit *Domain.BloodUnit) error
	UpdateDonorOverallStatus(donorID string, status string) error
	UpdateDonorBloodType(donorID string, bloodType string) error
	GetDonationByID(donationID string) (*Domain.DonationRecord, error)//
	GetTestResult(donationID string) (*Domain.DonorTestResult, error)
	GetPendingDonations() ([]Domain.DonationRecord, error)
    GetAllTestResults() ([]Domain.DonorTestResult, error)
    GetTestResultsByStatus(status string) ([]Domain.DonorTestResult, error)
    UpdateTestResult(result *Domain.DonorTestResult) error
	DeleteBloodUnit(donationID string) error
	 GetBloodUnitByDonationID(donationID string) (*Domain.BloodUnit, error)
    UpdateBloodUnit(unit *Domain.BloodUnit) error
	
}