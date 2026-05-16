package Domain

type SummaryAnalytics struct {
	TotalRequests  int `json:"total_requests"`
	TotalFulfilled int `json:"total_fulfilled"`
	TotalPending   int `json:"total_pending"`
	TotalCancelled int `json:"total_cancelled"`
}

type EmergencyAnalytics struct {
	TotalRequests  int `json:"total_requests"`
	TotalActive    int `json:"total_active"`
	TotalEnded     int `json:"total_ended"`
	TotalPublished int `json:"total_published"`
}

type InventoryAnalytics struct {
	TotalUnits     int `json:"total_units"`
	TotalAvailable int `json:"total_available"`
	TotalReserved  int `json:"total_reserved"`
	TotalUsed      int `json:"total_used"`
	TotalExpired   int `json:"total_expired"`
}

type ContractAnalytics struct {
	TotalContracts int `json:"total_contracts"`
	TotalActive    int `json:"total_active"`
	TotalPending   int `json:"total_pending"`
	TotalExpired   int `json:"total_expired"`
	TotalRejected  int `json:"total_rejected"`
}

type HospitalAnalyticsSummary struct {
	TotalHospitals int `json:"total_hospitals"`
	TotalActive    int `json:"total_active"`
	TotalPending   int `json:"total_pending"`
}
