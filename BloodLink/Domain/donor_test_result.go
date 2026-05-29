package Domain

import "time"

type DonorTestResult struct {
	TestID     string `json:"test_id"`
	DonationID string `json:"donation_id"`
	DonorID    string `json:"donor_id"`
	TestedBy   string `json:"tested_by"`

	// Resolved fields — populated in API responses, not stored in DB
	TesterName      string `json:"tester_name"`
	CampaignAddress string `json:"campaign_address"`
	DonationNumber  string `json:"donation_number"`

	HIVResult        string `json:"hiv_result"`
	HepatitisBResult string `json:"hepatitis_b_result"`
	HepatitisCResult string `json:"hepatitis_c_result"`
	SyphilisResult   string `json:"syphilis_result"`

	BloodType     string `json:"blood_type"`
	OverallStatus string `json:"overall_status"`

	CreatedAt time.Time `json:"created_at"`

	// --- Transient input fields (used in request body, NOT persisted in donor_test_results) ---
	StorageLocation string                `json:"storage_location,omitempty"`
	RackNumber      string                `json:"rack_number,omitempty"`
	ShelfNumber     string                `json:"shelf_number,omitempty"`
	Components      []BloodComponentInput `json:"components,omitempty"`
}

type TestResultListResponse struct {
	Total               int                `json:"total"`
	Cleared             int                `json:"cleared"`
	TemporarilyDeferred int                `json:"temporarily_deferred"`
	PermanentlyDeferred int                `json:"permanently_deferred"`
	Tests               []DonorTestResult  `json:"tests"`
}

type TestFilter struct {
	LabTechID       string
	OverallStatus   string
	BloodType       string
	ComponentType   string
	StorageLocation string
	DonationNumber  string
	StartDate       string
	EndDate         string
}