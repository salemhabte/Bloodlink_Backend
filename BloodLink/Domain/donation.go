package Domain

import "time"

// DonationRecord represents a blood donation process
type DonationRecord struct {
    DonationID     string    `json:"donation_id"`
    DonorID        string    `json:"donor_id"`
    CollectedBy    string    `json:"collected_by"`
    CollectionDate time.Time `json:"collection_date"`
    Weight         float64   `json:"weight"`
    BloodPressure  string    `json:"blood_pressure"`
    Hemoglobin     float64   `json:"hemoglobin"`
    Temperature    float64   `json:"temperature"`
    Pulse          int       `json:"pulse"`
    QuantityML     int       `json:"quantity_ml"`
    Status         string    `json:"status"`
    CreatedAt      time.Time `json:"created_at"`
}