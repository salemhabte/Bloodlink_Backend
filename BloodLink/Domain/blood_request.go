package Domain

import "time"

const (
	BloodRequestStatusPending            = "PENDING"
	BloodRequestStatusPartiallyFulfilled = "APPROVED_PARTIALLY_FULFILLED"
	BloodRequestStatusRejected           = "REJECTED"
	BloodRequestStatusFulfilled          = "FULFILLED"

	UrgencyLow      = "LOW"
	UrgencyMedium   = "MEDIUM"
	UrgencyHigh     = "HIGH"
	UrgencyCritical = "CRITICAL"
)

type BloodRequest struct {
	RequestID    string     `json:"request_id" db:"request_id"`
	HospitalID   string     `json:"hospital_id" db:"hospital_id"`
	BloodType    string     `json:"blood_type" db:"blood_type"`
	Component    string     `json:"component" db:"component"`
	Quantity     int        `json:"quantity" db:"quantity"`
	UrgencyLevel string     `json:"urgency_level" db:"urgency_level"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ApprovedAt   *time.Time `json:"approved_at" db:"approved_at"`
}

type BloodRequestItemDTO struct {
	BloodType string `json:"blood_type" binding:"required,oneof='A+' 'A-' 'B+' 'B-' 'AB+' 'AB-' 'O+' 'O-'"`
	Component string `json:"component" binding:"required,oneof=PRBC PLASMA PLATELETS CRYOPRECIPITATE CRYO_POOR_PLASMA WHOLE_BLOOD"`
	Quantity  int    `json:"quantity" binding:"required,gt=0"`
}

type CreateBloodRequestBatchDTO struct {
	HospitalID   string                `json:"hospital_id"`
	UrgencyLevel string                `json:"urgency_level" binding:"required,oneof=emergency normal"`
	Requests     []BloodRequestItemDTO `json:"requests" binding:"required,min=1,dive"`
}

type BloodRequestResponse struct {
	RequestID           string     `json:"request_id" db:"request_id"`
	HospitalID          string     `json:"hospital_id" db:"hospital_id"`
	HospitalName        string     `json:"hospital_name" db:"hospital_name"`
	BloodType           string     `json:"blood_type" db:"blood_type"`
	Component           string     `json:"component" db:"component"`
	Quantity            int        `json:"quantity" db:"quantity"`
	UrgencyLevel        string     `json:"urgency_level" db:"urgency_level"`
	Status              string     `json:"status" db:"status"`
	FulfilledCount      int        `json:"fulfilled_count" db:"fulfilled_count"`
	FulfilledQuantityMl int        `json:"fulfilled_quantity_ml" db:"fulfilled_quantity_ml"`
	Notes               string     `json:"notes" db:"notes"`
	IsUsed              bool       `json:"is_used" db:"is_used"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	ApprovedAt          *time.Time `json:"approved_at" db:"approved_at"`
}

type BloodRequestFilter struct {
	HospitalID   string `json:"hospital_id"`
	BloodType    string `json:"blood_type"`
	Component    string `json:"component"`
	Status       string `json:"status"`
	UrgencyLevel string `json:"urgency_level"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
}

type BloodRequestListResponse struct {
	Requests  []BloodRequestResponse `json:"requests"`
	Analytics SummaryAnalytics       `json:"analytics"`
}
