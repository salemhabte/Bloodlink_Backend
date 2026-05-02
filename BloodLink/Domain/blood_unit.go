package Domain

import "time"

type BloodUnit struct {
	BloodUnitID     string    `json:"blood_unit_id"`
	DonationID      string    `json:"donation_id"`
	BloodType       string    `json:"blood_type"`
	VolumeML        int       `json:"volume_ml"`
	CollectionDate  time.Time `json:"collection_date"`
	ExpirationDate  time.Time `json:"expiration_date"`
	Status          string    `json:"status"`
	ComponentType string `json:"component_type"`
	CreatedAt       time.Time `json:"created_at"`
}