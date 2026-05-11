package Domain

import "time"

type DonorProfile struct {
	FullName  string
	Email     string
	Phone     string
	Address   string
	BloodType string
}

type DonorBloodRequest struct {
	RequestID string
	DonorID   string

	DonorName    string
	DonorEmail   string
	DonorPhone   string
	DonorAddress string

	BloodType  string
	QuantityML int
	Reason     string

	HospitalName    string
	HospitalAddress string
	HospitalPhone   string

	Status    string
	CreatedAt time.Time
	
	// Resolved fields
	SuccessfulDonations int  `json:"successful_donations,omitempty"`
	CanFulfill          bool `json:"can_fulfill"`
}

type DonorBloodRequestFilter struct {
	StartDate string
	EndDate   string
	BloodType string
	Status    string
}