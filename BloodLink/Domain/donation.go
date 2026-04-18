package Domain

import "time"

// DonationRecord represents a blood donation process
type DonationRecord struct {
    DonationID     string    `json:"donation_id"`
    DonorID        string    `json:"donor_id"`
    CampaignID     *string   `json:"campaign_id"`
    CollectedBy    string    `json:"collected_by"`
    //It is donor name, collecter name and overall status
    //used only in the API response and not in the database
    DonorName      string    `json:"donor_name"`
    CollectorName  string    `json:"collector_name"`
    CollectionDate time.Time `json:"collection_date"`
    Weight         float64   `json:"weight"`
    BloodPressure  string    `json:"blood_pressure"`
    Hemoglobin     float64   `json:"hemoglobin"`
    Temperature    float64   `json:"temperature"`
    Pulse          int       `json:"pulse"`
    QuantityML     int       `json:"quantity_ml"`
    Status         string    `json:"status"`
    OverallStatus  string    `json:"overall_status"`
    CreatedAt      time.Time `json:"created_at"`
}