package Domain

import "time"

type HospitalDashboard struct {
	TotalRequests           int `json:"total_requests"`
	ApprovedRequests        int `json:"approved_requests"`
	PartiallyFulfilled      int `json:"partially_fulfilled"`
	RejectedRequests        int `json:"rejected_requests"`
	PendingRequests         int `json:"pending_requests"`
	
	ContractStatus          string    `json:"contract_status"`
	ContractEndDate         *time.Time `json:"contract_end_date"`
	
	MostRequestedBloodType  string    `json:"most_requested_blood_type"`
	TotalUnitsRequested     int       `json:"total_units_requested"`
	
	RecentRequests         []BloodRequestResponse `json:"recent_requests"`
	
	MonthlyRequestTrends   []MonthlyTrend `json:"monthly_request_trends"`
}

type MonthlyTrend struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}
