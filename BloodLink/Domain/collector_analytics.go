package Domain

type CollectorKPI struct {
	CollectorID string `json:"collector_id"`

	TotalDonations int `json:"total_donations"`

	TodaysDonations int `json:"todays_donations"`

	ApprovedCount int `json:"approved_count"`

	TemporarilyRejectedCount int `json:"temporarily_rejected_count"`

	ApprovalRate float64 `json:"approval_rate"`
	RejectionRate float64 `json:"rejection_rate"`
}
type CollectorDonorInsights struct {
	NewDonors       int `json:"new_donors"`
	ReturningDonors int `json:"returning_donors"`
	HighRiskDonors  int `json:"high_risk_donors"`
	MostActiveDonor string `json:"most_active_donor"`
}
type CollectorTodayStats struct {
	TodaysDonations int `json:"todays_donations"`
	PendingDonors   int `json:"pending_donors"`
	ApprovedToday   int `json:"approved_today"`
	RejectedToday   int `json:"rejected_today"`
}