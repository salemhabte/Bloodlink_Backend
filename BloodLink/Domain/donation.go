package Domain

import "time"

// DonationRecord represents a blood donation process
type DonationRecord struct {
    DonationID     string    `json:"donation_id"`
    DonorID        string    `json:"donor_id"`
    CampaignID     *string   `json:"campaign_id"`
    CollectedBy    string    `json:"collected_by"`

    // Resolved fields — used only in API responses, not stored in the database
    DonorName       string `json:"donor_name"`
    CollectorName   string `json:"collector_name"`
    CampaignTitle   string `json:"campaign_title"`
    CampaignAddress string `json:"campaign_address"`

    CollectionDate time.Time `json:"collection_date"`
    Weight         float64   `json:"weight"`
    BloodPressure  string    `json:"blood_pressure"`
    Hemoglobin     float64   `json:"hemoglobin"`
    Temperature    float64   `json:"temperature"`
    Pulse          int       `json:"pulse"`
    QuantityML     int       `json:"quantity_ml"`
    Status         string    `json:"status"`
    RejectionReason string   `json:"rejection_reason"`
    OverallStatus  string    `json:"overall_status"`
    CreatedAt      time.Time `json:"created_at"`
}

type DonationFilter struct {
	CollectorID string
	DonorID     string
	Status      string

	StartDate   string
	EndDate     string
}

type DonationListResponse struct {
	Total               int              `json:"total"`
	Approved            int              `json:"approved"`
	TemporarilyRejected int              `json:"temporarily_rejected"`
	Donations           []DonationRecord `json:"donations"`
}