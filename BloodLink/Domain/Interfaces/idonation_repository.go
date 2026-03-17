package Domain

import (
	domain "bloodlink/Domain"
)

type IDonationRepository interface {

	// CreateDonation inserts a new donation record into the database
	CreateDonation(record *domain.DonationRecord) error

	// SearchDonorByEmail finds a donor using their email
	SearchDonorByEmail(email string) (*domain.Donor, error)
	UpdateDonationStatus(donationID string, status string) error
}