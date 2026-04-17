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
	Quantity     int        `json:"quantity" db:"quantity"`
	UrgencyLevel string     `json:"urgency_level" db:"urgency_level"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ApprovedAt   *time.Time `json:"approved_at" db:"approved_at"`
}

type CreateBloodRequestDTO struct {
	BloodType    string `json:"blood_type" binding:"required,oneof='A+' 'A-' 'B+' 'B-' 'AB+' 'AB-' 'O+' 'O-'"`
	Quantity     int    `json:"quantity" binding:"required,gt=0"`
	UrgencyLevel string `json:"urgency_level" binding:"required,oneof=LOW MEDIUM HIGH CRITICAL"`
}

type UpdateBloodRequestStatusDTO struct {
	Status string `json:"status" binding:"required,oneof=PENDING APPROVED_PARTIALLY_FULFILLED REJECTED FULFILLED"`
}

type BloodRequestResponse struct {
	RequestID    string     `json:"request_id" db:"request_id"`
	HospitalID   string     `json:"hospital_id" db:"hospital_id"`
	HospitalName string     `json:"hospital_name" db:"hospital_name"`
	BloodType    string     `json:"blood_type" db:"blood_type"`
	Quantity     int        `json:"quantity" db:"quantity"`
	UrgencyLevel string     `json:"urgency_level" db:"urgency_level"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	ApprovedAt   *time.Time `json:"approved_at" db:"approved_at"`
}
