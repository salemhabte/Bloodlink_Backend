package Domain

import "time"

const (
	RequestStatusPending  = "PENDING"
	RequestStatusApproved = "APPROVED"
	RequestStatusRejected = "REJECTED"

	ContractStatusPending             = "PENDING"
	ContractStatusApprovedByHospital  = "APPROVED_BY_HOSPITAL"
	ContractStatusFinalized           = "FINALIZED"
	ContractStatusRejected            = "REJECTED"
)

type HospitalRequest struct {
	RequestID       string    `json:"request_id" db:"request_id"`
	HospitalName    string    `json:"hospital_name" db:"hospital_name"`
	Address         string    `json:"address" db:"address"`
	Phone           string    `json:"phone" db:"phone"`
	LicenseDocument string    `json:"license_document" db:"license_document"`
	Status          string    `json:"status" db:"status"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type HospitalRequestAdmin struct {
	RequestAdminID    string    `json:"request_admin_id" db:"request_admin_id"`
	RequestID         string    `json:"request_id" db:"request_id"`
	AdminFullName     string    `json:"admin_full_name" db:"admin_full_name"`
	AdminEmail        string    `json:"admin_email" db:"admin_email"`
	AdminPhone        string    `json:"admin_phone" db:"admin_phone"`
	AdminPasswordHash string    `json:"-" db:"admin_password_hash"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type Hospital struct {
	HospitalID string    `json:"hospital_id" db:"hospital_id"`
	Name       string    `json:"name" db:"name"`
	Address    string    `json:"address" db:"address"`
	Phone      string    `json:"phone" db:"phone"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type HospitalAdmin struct {
	HospitalAdminID string    `json:"hospital_admin_id" db:"hospital_admin_id"`
	UserID          string    `json:"user_id" db:"user_id"`
	HospitalID      string    `json:"hospital_id" db:"hospital_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type HospitalContract struct {
	ContractID            string    `json:"contract_id" db:"contract_id"`
	HospitalID            string    `json:"hospital_id" db:"hospital_id"`
	BloodBankAdminID      string    `json:"blood_bank_admin_id" db:"blood_bank_admin_id"`
	Document              *string   `json:"document" db:"document"`
	Status                string    `json:"status" db:"status"`
	ContractStart         *time.Time `json:"contract_start" db:"contract_start"`
	ContractEnd           *time.Time `json:"contract_end" db:"contract_end"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	HospitalSignaturePath *string   `json:"hospital_signature_path" db:"hospital_signature_path"`
	AdminSignaturePath    *string   `json:"admin_signature_path" db:"admin_signature_path"`
	TemplateID            *string   `json:"template_id" db:"template_id"`
}

// Request and Response DTOs
type RegisterHospitalRequestDTO struct {
	HospitalName  string `json:"hospital_name" binding:"required"`
	Address       string `json:"address" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	AdminFullName string `json:"admin_full_name" binding:"required"`
	AdminEmail    string `json:"admin_email" binding:"required,email"`
	AdminPhone    string `json:"admin_phone" binding:"required"`
	AdminPassword string `json:"admin_password" binding:"required,min=8"`
}

type SignContractRequestDTO struct {
	SignatureBase64 string `json:"signature_base64" binding:"required"`
}

type ApproveHospitalRequestDTO struct {
	TemplateID string `json:"template_id" binding:"required"`
}

type ContractTemplate struct {
	TemplateID string    `json:"template_id" db:"template_id"`
	Name       string    `json:"name" db:"name"`
	Content    string    `json:"content" db:"content"`
	CreatedBy  *string   `json:"created_by" db:"created_by"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type CreateTemplateRequestDTO struct {
	Name    string `json:"name" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type HospitalRequestResponse struct {
	RequestID    string `json:"request_id"`
	HospitalName string `json:"hospital_name"`
	Address      string `json:"address"`
	Phone        string `json:"phone"`
	Status       string `json:"status"`
	AdminName    string `json:"admin_name"`
	AdminEmail   string `json:"admin_email"`
}
