package Domain

import "time"

type DonorTestResult struct {
	TestID      string `json:"test_id"`
	DonationID  string `json:"donation_id"`
	DonorID     string `json:"donor_id"`
	TestedBy    string `json:"tested_by"`

	// Resolved fields — populated in API responses, not stored in DB
	TesterName      string `json:"tester_name"`
	CampaignAddress string `json:"campaign_address"`

	HIVResult       string `json:"hiv_result"`
	HepatitisResult string `json:"hepatitis_result"`
	SyphilisResult  string `json:"syphilis_result"`

	BloodType     string `json:"blood_type"`
	ComponentType string `json:"component_type"`

	OverallStatus string `json:"overall_status"`

	CreatedAt time.Time `json:"created_at"`
}