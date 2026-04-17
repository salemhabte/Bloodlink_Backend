package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
)

type hospitalRepository struct {
	db *sql.DB
}

func NewHospitalRepository(db *sql.DB) Interfaces.IHospitalRepository {
	return &hospitalRepository{db: db}
}

func (r *hospitalRepository) CreateHospitalRequest(req *Domain.HospitalRequest) error {
	query := `INSERT INTO hospital_requests (request_id, hospital_name, address, phone, license_document, status, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, req.RequestID, req.HospitalName, req.Address, req.Phone, req.LicenseDocument, req.Status, req.CreatedAt)
	return err
}

func (r *hospitalRepository) CreateHospitalRequestAdmin(admin *Domain.HospitalRequestAdmin) error {
	query := `INSERT INTO hospital_request_admins (request_admin_id, request_id, admin_full_name, admin_email, admin_phone, admin_password_hash, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, admin.RequestAdminID, admin.RequestID, admin.AdminFullName, admin.AdminEmail, admin.AdminPhone, admin.AdminPasswordHash, admin.CreatedAt)
	return err
}

func (r *hospitalRepository) GetPendingRequests() ([]Domain.HospitalRequestResponse, error) {
	query := `SELECT r.request_id, r.hospital_name, r.address, r.phone, r.status, a.admin_full_name, a.admin_email
			  FROM hospital_requests r
			  JOIN hospital_request_admins a ON r.request_id = a.request_id
			  WHERE r.status = $1`
	rows, err := r.db.Query(query, Domain.RequestStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Domain.HospitalRequestResponse
	for rows.Next() {
		var req Domain.HospitalRequestResponse
		if err := rows.Scan(&req.RequestID, &req.HospitalName, &req.Address, &req.Phone, &req.Status, &req.AdminName, &req.AdminEmail); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *hospitalRepository) GetHospitalRequestByID(requestID string) (*Domain.HospitalRequest, *Domain.HospitalRequestAdmin, error) {
	reqQuery := `SELECT request_id, hospital_name, address, phone, license_document, status, created_at FROM hospital_requests WHERE request_id = $1`
	req := &Domain.HospitalRequest{}
	err := r.db.QueryRow(reqQuery, requestID).Scan(&req.RequestID, &req.HospitalName, &req.Address, &req.Phone, &req.LicenseDocument, &req.Status, &req.CreatedAt)
	if err != nil {
		return nil, nil, err
	}

	adminQuery := `SELECT request_admin_id, request_id, admin_full_name, admin_email, admin_phone, admin_password_hash, created_at FROM hospital_request_admins WHERE request_id = $1`
	admin := &Domain.HospitalRequestAdmin{}
	err = r.db.QueryRow(adminQuery, requestID).Scan(&admin.RequestAdminID, &admin.RequestID, &admin.AdminFullName, &admin.AdminEmail, &admin.AdminPhone, &admin.AdminPasswordHash, &admin.CreatedAt)
	if err != nil {
		return nil, nil, err
	}

	return req, admin, nil
}

func (r *hospitalRepository) UpdateHospitalRequestStatus(requestID string, status string) error {
	query := `UPDATE hospital_requests SET status = $1 WHERE request_id = $2`
	_, err := r.db.Exec(query, status, requestID)
	return err
}

func (r *hospitalRepository) CreateHospital(hospital *Domain.Hospital) error {
	query := `INSERT INTO hospitals (hospital_id, name, address, phone, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, hospital.HospitalID, hospital.Name, hospital.Address, hospital.Phone, hospital.CreatedAt)
	return err
}

func (r *hospitalRepository) CreateHospitalAdmin(admin *Domain.HospitalAdmin) error {
	query := `INSERT INTO hospital_admins (hospital_admin_id, user_id, hospital_id, created_at)
			  VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, admin.HospitalAdminID, admin.UserID, admin.HospitalID, admin.CreatedAt)
	return err
}

func (r *hospitalRepository) GetHospitalAdminByUserID(userID string) (*Domain.HospitalAdmin, error) {
	query := `SELECT hospital_admin_id, user_id, hospital_id, created_at FROM hospital_admins WHERE user_id = $1`
	admin := &Domain.HospitalAdmin{}
	err := r.db.QueryRow(query, userID).Scan(&admin.HospitalAdminID, &admin.UserID, &admin.HospitalID, &admin.CreatedAt)
	return admin, err
}

func (r *hospitalRepository) CreateContract(contract *Domain.HospitalContract) error {
	query := `INSERT INTO hospital_contracts (contract_id, hospital_id, blood_bank_admin_id, document, status, contract_start, contract_end, created_at, template_id)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(query, contract.ContractID, contract.HospitalID, contract.BloodBankAdminID, contract.Document, contract.Status, contract.ContractStart, contract.ContractEnd, contract.CreatedAt, contract.TemplateID)
	return err
}

func (r *hospitalRepository) GetContractByID(contractID string) (*Domain.HospitalContract, error) {
	query := `SELECT contract_id, hospital_id, blood_bank_admin_id, document, status, contract_start, contract_end, created_at, hospital_signature_path, admin_signature_path, template_id
			  FROM hospital_contracts WHERE contract_id = $1`
	contract := &Domain.HospitalContract{}
	err := r.db.QueryRow(query, contractID).Scan(&contract.ContractID, &contract.HospitalID, &contract.BloodBankAdminID, &contract.Document, &contract.Status, &contract.ContractStart, &contract.ContractEnd, &contract.CreatedAt, &contract.HospitalSignaturePath, &contract.AdminSignaturePath, &contract.TemplateID)
	return contract, err
}

func (r *hospitalRepository) GetHospitalByID(hospitalID string) (*Domain.Hospital, error) {
	query := `SELECT hospital_id, name, address, phone, created_at FROM hospitals WHERE hospital_id = $1`
	hospital := &Domain.Hospital{}
	err := r.db.QueryRow(query, hospitalID).Scan(&hospital.HospitalID, &hospital.Name, &hospital.Address, &hospital.Phone, &hospital.CreatedAt)
	return hospital, err
}

func (r *hospitalRepository) UpdateContract(contract *Domain.HospitalContract) error {
	query := `UPDATE hospital_contracts SET status = $1, document = $2, hospital_signature_path = $3, admin_signature_path = $4 WHERE contract_id = $5`
	_, err := r.db.Exec(query, contract.Status, contract.Document, contract.HospitalSignaturePath, contract.AdminSignaturePath, contract.ContractID)
	return err
}

func (r *hospitalRepository) CreateContractTemplate(template *Domain.ContractTemplate) error {
	query := `INSERT INTO contract_templates (template_id, name, content, created_by, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, template.TemplateID, template.Name, template.Content, template.CreatedBy, template.CreatedAt)
	return err
}

func (r *hospitalRepository) GetContractTemplates() ([]Domain.ContractTemplate, error) {
	query := `SELECT template_id, name, content, created_by, created_at FROM contract_templates`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []Domain.ContractTemplate
	for rows.Next() {
		var t Domain.ContractTemplate
		if err := rows.Scan(&t.TemplateID, &t.Name, &t.Content, &t.CreatedBy, &t.CreatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *hospitalRepository) GetContractTemplateByID(templateID string) (*Domain.ContractTemplate, error) {
	query := `SELECT template_id, name, content, created_by, created_at FROM contract_templates WHERE template_id = $1`
	t := &Domain.ContractTemplate{}
	err := r.db.QueryRow(query, templateID).Scan(&t.TemplateID, &t.Name, &t.Content, &t.CreatedBy, &t.CreatedAt)
	return t, err
}

func (r *hospitalRepository) UpdateContractTemplate(template *Domain.ContractTemplate) error {
	query := `UPDATE contract_templates SET name = $1, content = $2 WHERE template_id = $3`
	_, err := r.db.Exec(query, template.Name, template.Content, template.TemplateID)
	return err
}

func (r *hospitalRepository) DeleteContractTemplate(templateID string) error {
	query := `DELETE FROM contract_templates WHERE template_id = $1`
	_, err := r.db.Exec(query, templateID)
	return err
}
