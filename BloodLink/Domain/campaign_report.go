package Domain

type CampaignStats struct {
	TotalCampaigns   int `json:"total_campaigns"`
	ActiveCampaigns  int `json:"active_campaigns"`
	CompletedCampaigns int `json:"completed_campaigns"`
	UpcomingCampaigns int `json:"upcoming_campaigns"`
}

type CampaignDonationStats struct {
	CampaignID        string  `json:"campaign_id"`
	TotalDonations    int     `json:"total_donations"`
	TotalBloodML      int     `json:"total_blood_ml"`
	AvgDonationPerDonor float64 `json:"avg_donation_per_donor"`
	ApprovedCount     int     `json:"approved_count"`
	TemporarilyRejectedCount int `json:"temporarily_rejected_count"`
	SuccessRate       float64 `json:"success_rate"`
	DropOffRate       float64 `json:"dropoff_rate"`
}