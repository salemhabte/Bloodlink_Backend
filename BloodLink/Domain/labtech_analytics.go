package Domain

type LabDashboard struct {
	TotalTests int `json:"total_tests"`

	PendingTests int `json:"pending_tests"` // GLOBAL (from donation_records)

	Cleared int `json:"cleared"`
	TemporarilyDeferred int `json:"temporarily_deferred"`
	PermanentlyDeferred int `json:"permanently_deferred"`

	HIVPositive int `json:"hiv_positive"`
	HepatitisPositive int `json:"hepatitis_positive"`
	SyphilisPositive int `json:"syphilis_positive"`

	ClearedPercent float64 `json:"cleared_percent"`
	TemporarilyDeferredPercent float64 `json:"temporarily_deferred_percent"`
	PermanentlyDeferredPercent float64 `json:"permanently_deferred_percent"`
}
