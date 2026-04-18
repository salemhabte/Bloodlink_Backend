package Domain

import (
	domain "bloodlink/Domain"
)

type IDonationRepository interface {

	// CreateDonation inserts a new donation record into the database
	CreateDonation(record *domain.DonationRecord) error

	// Search donor by email or phone
	SearchDonor(query string) (*domain.DonorResponse, error)
	UpdateDonationStatus(donationID string, status string) error
	// Update donation medical info
	UpdateDonation(record *domain.DonationRecord) error

	GetDonationByID(id string) (*domain.DonationRecord, error)
	// Get all donations
	GetAllDonations() ([]domain.DonationRecord, error)
	// Get last donation for 3 month rule
	GetLastDonationByDonor(donorID string) (*domain.DonationRecord, error)
	UpdateDonorWeight(donorID string, weight float64) error
GetPendingDonors() ([]domain.DonorResponse, error)
GetPendingDonorByID(donorID string) (*domain.DonorResponse, error)
SearchPendingDonor(query string) (*domain.DonorResponse, error)
GetAllDonationsByDonor(donorID string) ([]domain.DonationRecord, error)
}