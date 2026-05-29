package Domain

import (
	"bloodlink/Domain"
	"time"
)
type IDonorBloodRequestRepository interface {

	// ===== CRUD =====
	Create(req *Domain.DonorBloodRequest) error
	GetAllAdmin(filter Domain.DonorBloodRequestFilter) ([]Domain.DonorBloodRequest, error)
	GetByID(id string) (*Domain.DonorBloodRequest, error)
	GetByDonorID(donorID string, filter Domain.DonorBloodRequestFilter) ([]Domain.DonorBloodRequest, error)
	UpdateStatus(id string, status string) error
	UpdateStatusWithUnits(id string, status string, reservedUnits int) error

	// ===== Donor Info =====
	GetDonorIDByUserID(userID string) (string, error)
	GetDonorProfile(donorID string) (*Domain.DonorProfile, error)

	// ===== Blood Inventory =====
	GetAvailableBloodUnits(bloodType string) ([]string, error)
	ReserveBloodUnits(requestID string, bloodType string, componentType string, requiredUnits int) (int, error)
	MarkReservedAsUsed(requestID string) error
	ExpireStaleReservations() error

	// ===== Validation =====
	HasSuccessfulDonation(donorID string) (bool, error)
	IsDonorInTop10(donorID string) (bool, error)
	GetLastRequestDateByDonor(donorID string) (time.Time, error)
}