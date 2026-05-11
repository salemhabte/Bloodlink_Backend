package Domain

import "time"

type BloodUnit struct {
	BloodUnitID   string    `json:"blood_unit_id"`
	DonationID    string    `json:"donation_id"`
	BloodType     string    `json:"blood_type"`
	ComponentType string    `json:"component_type"`
	VolumeML      int       `json:"volume_ml"`
	CollectionDate time.Time `json:"collection_date"`
	ExpirationDate time.Time `json:"expiration_date"`
	Status        string    `json:"status"` // AVAILABLE | RESERVED | USED | EXPIRED

	// Reservation fields (populated when status = RESERVED)
	ReservedForHospitalID string     `json:"reserved_for_hospital_id,omitempty"`
	ReservedAt            *time.Time `json:"reserved_at,omitempty"`
	RequestID             string     `json:"request_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

type BloodUnitFilter struct {
	BloodType     string `json:"blood_type"`
	ComponentType string `json:"component_type"`
	Status        string `json:"status"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
}

// ReservedUnitInfo is returned when units are reserved for a hospital request
type ReservedUnitInfo struct {
	BloodUnitID    string    `json:"blood_unit_id"`
	BloodType      string    `json:"blood_type"`
	VolumeML       int       `json:"volume_ml"`
	ExpirationDate time.Time `json:"expiration_date"`
}

// ApproveRequestResult is the response returned when admin approves a blood request
type ApproveRequestResult struct {
	Status         string             `json:"status"`
	Message        string             `json:"message"`
	ReservedUnits  []ReservedUnitInfo `json:"reserved_units"`
	TotalVolumeML  int                `json:"total_volume_ml"`
	RequestedCount int                `json:"requested_count"`
	FulfilledCount int                `json:"fulfilled_count"`
}