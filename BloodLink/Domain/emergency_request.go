package Domain

import "time"

const (
	EmergencyStatusPending   = "PENDING_PUBLISH"
	EmergencyStatusPublished = "PUBLISHED"
	EmergencyStatusRejected  = "REJECTED"
	EmergencyStatusCompleted = "COMPLETED"
)

type EmergencyRequest struct {
	EmergencyID       string     `json:"emergency_id" db:"emergency_id"`
	RequestID         *string    `json:"request_id,omitempty" db:"request_id"`
	BloodType         string     `json:"blood_type" db:"blood_type"`
	QuantityRequired  int        `json:"quantity_required" db:"quantity_required"`
	QuantityFulfilled int        `json:"quantity_fulfilled" db:"quantity_fulfilled"`
	UrgencyLevel      string     `json:"urgency_level" db:"urgency_level"`
	HospitalName      string     `json:"hospital_name" db:"hospital_name"`
	Location          string     `json:"location" db:"location"`
	Status            string     `json:"status" db:"status"`
	IsManual          bool       `json:"is_manual" db:"is_manual"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	PublishedAt       *time.Time `json:"published_at,omitempty" db:"published_at"`
	EndDate           *time.Time `json:"end_date,omitempty" db:"end_date"`
	Latitude          float64    `json:"latitude" db:"latitude"`
	Longitude         float64    `json:"longitude" db:"longitude"`
}

type CreateEmergencyRequestDTO struct {
	BloodType        string  `json:"blood_type" binding:"required,oneof='A+' 'A-' 'B+' 'B-' 'AB+' 'AB-' 'O+' 'O-'"`
	QuantityRequired int     `json:"quantity_required" binding:"omitempty,gt=0"`
	HospitalName     string  `json:"hospital_name" binding:"required"`
	Location         string  `json:"location" binding:"required"`
	EndDate          string  `json:"end_date" binding:"required"`
	Latitude         float64 `json:"latitude,string" binding:"required"`
	Longitude        float64 `json:"longitude,string" binding:"required"`
}

type EmergencyRequestFilter struct {
	BloodType    string `json:"blood_type"`
	Status       string `json:"status"`
	UrgencyLevel string `json:"urgency_level"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
}

type EmergencyListResponse struct {
	Emergencies []EmergencyRequest `json:"emergencies"`
	Analytics   EmergencyAnalytics `json:"analytics"`
}
