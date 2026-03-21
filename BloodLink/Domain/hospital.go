package Domain

import "time"

// Hospital represents the database model for a hospital
type Hospital struct {
	HospitalID         string    `json:"hospital_id" db:"hospital_id"`
	HospitalName       string    `json:"hospital_name" db:"hospital_name"`
	Address            string    `json:"address" db:"address"`
	City               string    `json:"city" db:"city"`
	Phone              string    `json:"phone" db:"phone"`
	ContactPersonName  string    `json:"contact_person_name" db:"contact_person_name"`
	ContactPersonPhone string    `json:"contact_person_phone" db:"contact_person_phone"`
	Status             string    `json:"status" db:"status"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

// RegisterHospitalRequest DTO
type RegisterHospitalRequest struct {
	HospitalName       string `json:"hospital_name" binding:"required"`
	Address            string `json:"address" binding:"required"`
	City               string `json:"city" binding:"required"`
	Phone              string `json:"phone" binding:"required"`
	ContactPersonName  string `json:"contact_person_name" binding:"required"`
	ContactPersonPhone string `json:"contact_person_phone" binding:"required"`
}

// UpdateHospitalRequest DTO
type UpdateHospitalRequest struct {
	HospitalName       string `json:"hospital_name"`
	Address            string `json:"address"`
	City               string `json:"city"`
	Phone              string `json:"phone"`
	ContactPersonName  string `json:"contact_person_name"`
	ContactPersonPhone string `json:"contact_person_phone"`
	Status             string `json:"status"`
}
