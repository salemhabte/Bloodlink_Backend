package Usecase

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
)

// DonationUsecase contains business logic for blood donations
type DonationUsecase struct {
	repo Interfaces.IDonationRepository
}

// NewDonationUsecase creates a new instance of DonationUsecase
func NewDonationUsecase(repo Interfaces.IDonationRepository) *DonationUsecase {
	return &DonationUsecase{repo: repo}
}

// CreateDonation handles the business logic for recording a new donation
func (u *DonationUsecase) CreateDonation(record *Domain.DonationRecord) error {
	// Add business logic here (e.g., validation)
	return u.repo.CreateDonation(record)
}

// SearchDonorByEmail handles finding a donor by their email address
func (u *DonationUsecase) SearchDonorByEmail(email string) (*Domain.Donor, error) {
	return u.repo.SearchDonorByEmail(email)
}

// UpdateDonationStatus handles updating the status of a donation
func (u *DonationUsecase) UpdateDonationStatus(donationID string, status string) error {
	// You can add validation logic here if needed
	return u.repo.UpdateDonationStatus(donationID, status)
}