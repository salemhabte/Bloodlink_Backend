package Domain

// ================= SUPPORTING TYPES =================
type CollectorDonationStats struct {
	CollectorID    string `json:"collectorId"`
	TotalDonations int    `json:"totalDonations"`
}

type BloodTypeStat struct {
	BloodType string  `json:"bloodType"`
	Count     int     `json:"count"`
	Percent   float64 `json:"percent"`
}

// ================= DONOR RESPONSE =================
type DonorSummaryResponse struct {
	TotalRegisteredDonors     int `json:"totalRegisteredDonors"`
	DonorsWithDonationRecord  int `json:"donorsWithDonationRecord"`
	ApprovedDonors            int `json:"approvedDonors"`
	TemporarilyRejectedDonors int `json:"temporarilyRejectedDonors"`
}

// ================= SCREENING RESPONSE =================
type ScreeningSummaryResponse struct {
	ClearedDonors            int     `json:"clearedDonors"`
	TemporarilyDeferred      int     `json:"temporarilyDeferred"`
	PermanentlyDeferred      int     `json:"permanentlyDeferred"`
	ClearedPercent           float64 `json:"clearedPercent"`
	TempDeferredPercent      float64 `json:"tempDeferredPercent"`
	PermanentDeferredPercent float64 `json:"permanentDeferredPercent"`
}

// ================= COLLECTOR RESPONSE =================
type CollectorSummaryResponse struct {
	TotalCollectors      int                       `json:"totalCollectors"`
	TotalDonationRecords int                       `json:"totalDonationRecords"`
	DonationPerCollector []CollectorDonationStats `json:"donationPerCollector"`
}

// ================= LAB RESPONSE =================
type LabSummaryResponse struct {
	TotalLabTechs     int            `json:"totalLabTechs"`
	TotalTestResults  int            `json:"totalTestResults"`

	LabCleared        int     `json:"labCleared"`
	LabTempDeferred   int     `json:"labTempDeferred"`
	LabPermDeferred   int     `json:"labPermDeferred"`

	LabClearedPercent float64 `json:"labClearedPercent"`
	LabTempPercent    float64 `json:"labTempPercent"`
	LabPermPercent    float64 `json:"labPermPercent"`

	TestsPerLabTech   []LabTestStats `json:"testsPerLabTech"` // 
}

// ================= INVENTORY RESPONSE =================
type InventorySummaryResponse struct {
	TotalBloodUnits int             `json:"totalBloodUnits"`
	BloodTypeStats  []BloodTypeStat `json:"bloodTypeStats"`
	NearExpiryUnits int             `json:"nearExpiryUnits"`
}

// ================= HOSPITAL RESPONSE =================
type HospitalSummaryResponse struct {
	TotalHospitals          int `json:"totalHospitals"`
	ActiveContracts         int `json:"activeContracts"`
	PendingHospitalRequests int `json:"pendingHospitalRequests"`
	ActiveEmergencies       int `json:"activeEmergencies"`
}

type AdminDashboard struct {
	TotalRegisteredDonors     int `json:"totalRegisteredDonors"`
	DonorsWithDonationRecord  int `json:"donorsWithDonationRecord"`
	ApprovedDonors            int `json:"approvedDonors"`
	TemporarilyRejectedDonors int `json:"temporarilyRejectedDonors"`

	ClearedDonors       int `json:"clearedDonors"`
	TemporarilyDeferred int `json:"temporarilyDeferred"`
	PermanentlyDeferred int `json:"permanentlyDeferred"`

	ClearedPercent           float64 `json:"clearedPercent"`
	TempDeferredPercent      float64 `json:"tempDeferredPercent"`
	PermanentDeferredPercent float64 `json:"permanentDeferredPercent"`

	TotalCollectors      int                       `json:"totalCollectors"`
	TotalDonationRecords int                       `json:"totalDonationRecords"`
	DonationPerCollector []CollectorDonationStats `json:"donationPerCollector"`

	TotalLabTechs    int            `json:"totalLabTechs"`
	TotalTestResults int            `json:"totalTestResults"`
	LabCleared       int            `json:"labCleared"`
	LabTempDeferred  int            `json:"labTempDeferred"`
	LabPermDeferred  int            `json:"labPermDeferred"`

	LabClearedPercent float64        `json:"labClearedPercent"`
	LabTempPercent    float64        `json:"labTempPercent"`
	LabPermPercent    float64        `json:"labPermPercent"`
	TestsPerLabTech   []LabTestStats `json:"testsPerLabTech"`

	TotalBloodUnits int             `json:"totalBloodUnits"`
	BloodTypeStats  []BloodTypeStat `json:"bloodTypeStats"`
	NearExpiryUnits int             `json:"nearExpiryUnits"`

	// HOSPITAL METRICS
	TotalHospitals          int `json:"totalHospitals"`
	ActiveContracts         int `json:"activeContracts"`
	PendingHospitalRequests int `json:"pendingHospitalRequests"`
	ActiveEmergencies       int `json:"activeEmergencies"`
}

type LabTestStats struct {
	LabTechID  string `json:"labTechId"`
	TotalTests int    `json:"totalTests"`
}
