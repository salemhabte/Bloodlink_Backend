package Domain

import "time"

// DonationRecord represents a blood donation process
type DonationRecord struct {

	DonationID string

	DonorID string

	CollectedBy string

	CollectionDate time.Time

	Weight float64

	BloodPressure string

	Hemoglobin float64

	Temperature float64

	Pulse int

	QuantityML int

	Status string

	CreatedAt time.Time
}