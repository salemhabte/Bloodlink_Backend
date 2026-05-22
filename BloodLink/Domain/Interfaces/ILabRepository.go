package Domain

import "bloodlink/Domain"

type ILabRepository interface {
	CreateTestResult(result *Domain.DonorTestResult) error
	CreateBloodUnit(unit *Domain.BloodUnit) error
	UpdateDonorOverallStatus(donorID string, status string) error
	UpdateDonorBloodType(donorID string, bloodType string) error
	GetDonationByID(donationID string) (*Domain.DonationRecord, error)
	GetPendingDonationByID(donationID string) (*Domain.DonationRecord, error)
	GetTestResult(donationID string) (*Domain.DonorTestResult, error)
	GetPendingDonations() ([]Domain.DonationRecord, error)
	GetAllTestResults() ([]Domain.DonorTestResult, error)
	GetTestResultsByStatus(status string) ([]Domain.DonorTestResult, error)
	UpdateTestResult(result *Domain.DonorTestResult) error
	DeleteBloodUnit(donationID string) error
	DeleteBloodUnitsByDonationID(donationID string) error
	GetBloodUnitByDonationID(donationID string) (*Domain.BloodUnit, error)
	GetBloodUnitsByDonationID(donationID string) ([]Domain.BloodUnit, error)
	UpdateBloodUnit(unit *Domain.BloodUnit) error
	GetTestResultsByLabTech(labTechID string) ([]Domain.DonorTestResult, error)
	FilterTestResults(filter Domain.TestFilter) ([]Domain.DonorTestResult, error)
	GetMyTestResultsFiltered(filter Domain.TestFilter) ([]Domain.DonorTestResult, error)
	GetLatestTestResultByDonor(donorID string) (*Domain.DonorTestResult, error)
	IsSlotOccupied(location, rack, shelf, position string) (bool, error)
	GetOccupiedSlotCount(location, rack, shelf string) (int, error)
}