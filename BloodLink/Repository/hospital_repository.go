package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
	"fmt"
)

type hospitalRepository struct {
	db *sql.DB
}

func NewHospitalRepository(db *sql.DB) Interfaces.IHospitalRepository {
	return &hospitalRepository{db: db}
}

func (r *hospitalRepository) CreateHospitalRequest(req *Domain.HospitalRequest) error {
	query := `INSERT INTO hospital_requests (request_id, hospital_name, address, phone, license_document, status, created_at, latitude, longitude, location_geo)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, ST_SetSRID(ST_MakePoint($9, $8), 4326)::geography)`
	_, err := r.db.Exec(query, req.RequestID, req.HospitalName, req.Address, req.Phone, req.LicenseDocument, req.Status, req.CreatedAt, req.Latitude, req.Longitude)
	return err
}

func (r *hospitalRepository) CreateHospitalRequestAdmin(admin *Domain.HospitalRequestAdmin) error {
	query := `INSERT INTO hospital_request_admins (request_admin_id, request_id, admin_full_name, admin_email, admin_phone, admin_password_hash, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, admin.RequestAdminID, admin.RequestID, admin.AdminFullName, admin.AdminEmail, admin.AdminPhone, admin.AdminPasswordHash, admin.CreatedAt)
	return err
}

func (r *hospitalRepository) CreateHospitalRegistrationRequest(req *Domain.HospitalRequest, admin *Domain.HospitalRequestAdmin) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// 1. Insert Hospital Request
	queryReq := `INSERT INTO hospital_requests (request_id, hospital_name, address, phone, license_document, status, created_at, latitude, longitude, location_geo)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, ST_SetSRID(ST_MakePoint($9, $8), 4326)::geography)`
	_, err = tx.Exec(queryReq, req.RequestID, req.HospitalName, req.Address, req.Phone, req.LicenseDocument, req.Status, req.CreatedAt, req.Latitude, req.Longitude)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Insert Request Admin
	queryAdmin := `INSERT INTO hospital_request_admins (request_admin_id, request_id, admin_full_name, admin_email, admin_phone, admin_password_hash, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(queryAdmin, admin.RequestAdminID, admin.RequestID, admin.AdminFullName, admin.AdminEmail, admin.AdminPhone, admin.AdminPasswordHash, admin.CreatedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *hospitalRepository) ApproveHospitalRegistration(hospital *Domain.Hospital, user *Domain.User, admin *Domain.HospitalAdmin, contract *Domain.HospitalContract, requestID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// 1. Create real Hospital
	queryHospital := `INSERT INTO hospitals (hospital_id, name, address, phone, created_at, latitude, longitude, location_geo, license_document)
					  VALUES ($1, $2, $3, $4, $5, $6, $7, ST_SetSRID(ST_MakePoint($7, $6), 4326)::geography, $8)`
	_, err = tx.Exec(queryHospital, hospital.HospitalID, hospital.Name, hospital.Address, hospital.Phone, hospital.CreatedAt, hospital.Latitude, hospital.Longitude, hospital.LicenseDocument)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Create actual User
	queryUser := `INSERT INTO users (user_id, email, full_name, phone, password_hash, role, is_active, created_at)
				  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = tx.Exec(queryUser, user.ID, user.Email, user.FullName, user.Phone, user.Password, user.Role, user.IsActive, user.CreatedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 3. Create Hospital Admin record
	queryAdmin := `INSERT INTO hospital_admins (hospital_admin_id, user_id, hospital_id, created_at)
				   VALUES ($1, $2, $3, $4)`
	_, err = tx.Exec(queryAdmin, admin.HospitalAdminID, admin.UserID, admin.HospitalID, admin.CreatedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 4. Create Contract Record
	queryContract := `INSERT INTO hospital_contracts (contract_id, hospital_id, blood_bank_admin_id, status, document, contract_start, contract_end, created_at, template_id)
					  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = tx.Exec(queryContract, contract.ContractID, contract.HospitalID, contract.BloodBankAdminID, contract.Status, contract.Document, contract.ContractStart, contract.ContractEnd, contract.CreatedAt, contract.TemplateID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 5. Update Request Status to Approved
	queryRequest := `UPDATE hospital_requests SET status = $1 WHERE request_id = $2`
	_, err = tx.Exec(queryRequest, Domain.RequestStatusApproved, requestID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *hospitalRepository) GetPendingRequests(filter Domain.HospitalRequestFilter) ([]Domain.HospitalRequestResponse, error) {
	query := `SELECT r.request_id, r.hospital_name, r.address, r.phone, r.status, a.admin_full_name, a.admin_email, r.license_document
			  FROM hospital_requests r
			  JOIN hospital_request_admins a ON r.request_id = a.request_id
			  WHERE 1=1`

	args := []interface{}{}
	placeholderID := 1

	status := filter.Status
	if status == "" {
		status = Domain.RequestStatusPending
	}
	query += fmt.Sprintf(" AND r.status = $%d", placeholderID)
	args = append(args, status)
	placeholderID++

	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND r.created_at >= $%d", placeholderID)
		args = append(args, filter.StartDate)
		placeholderID++
	}
	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND r.created_at <= $%d", placeholderID)
		args = append(args, filter.EndDate)
		placeholderID++
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Domain.HospitalRequestResponse
	for rows.Next() {
		var req Domain.HospitalRequestResponse
		if err := rows.Scan(&req.RequestID, &req.HospitalName, &req.Address, &req.Phone, &req.Status, &req.AdminName, &req.AdminEmail, &req.LicenseDocument); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *hospitalRepository) GetHospitalRequestByID(requestID string) (*Domain.HospitalRequest, *Domain.HospitalRequestAdmin, error) {
	reqQuery := `SELECT request_id, hospital_name, address, phone, license_document, status, created_at, latitude, longitude FROM hospital_requests WHERE request_id = $1`
	req := &Domain.HospitalRequest{}
	err := r.db.QueryRow(reqQuery, requestID).Scan(&req.RequestID, &req.HospitalName, &req.Address, &req.Phone, &req.LicenseDocument, &req.Status, &req.CreatedAt, &req.Latitude, &req.Longitude)
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
	query := `INSERT INTO hospitals (hospital_id, name, address, phone, created_at, latitude, longitude, location_geo, license_document) VALUES ($1, $2, $3, $4, $5, $6, $7, ST_SetSRID(ST_MakePoint($7, $6), 4326)::geography, $8)`
	_, err := r.db.Exec(query, hospital.HospitalID, hospital.Name, hospital.Address, hospital.Phone, hospital.CreatedAt, hospital.Latitude, hospital.Longitude, hospital.LicenseDocument)
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
	query := `SELECT c.contract_id, c.hospital_id, h.name as hospital_name, c.blood_bank_admin_id, c.document, 
			         c.status, c.contract_start, c.contract_end, c.created_at, c.hospital_signature_path, c.admin_signature_path, c.template_id, COALESCE(h.license_document, '')
			  FROM hospital_contracts c
			  JOIN hospitals h ON c.hospital_id = h.hospital_id
			  WHERE c.contract_id = $1`
	contract := &Domain.HospitalContract{}
	var hName string
	var licenseDoc string
	err := r.db.QueryRow(query, contractID).Scan(&contract.ContractID, &contract.HospitalID, &hName, &contract.BloodBankAdminID, &contract.Document, &contract.Status, &contract.ContractStart, &contract.ContractEnd, &contract.CreatedAt, &contract.HospitalSignaturePath, &contract.AdminSignaturePath, &contract.TemplateID, &licenseDoc)
	return contract, err
}

func (r *hospitalRepository) GetContractsByHospitalID(hospitalID string) ([]Domain.HospitalContract, error) {
	query := `SELECT contract_id, hospital_id, blood_bank_admin_id, document, status, contract_start, contract_end, created_at, hospital_signature_path, admin_signature_path, template_id
			  FROM hospital_contracts WHERE hospital_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(query, hospitalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []Domain.HospitalContract
	for rows.Next() {
		var c Domain.HospitalContract
		err := rows.Scan(&c.ContractID, &c.HospitalID, &c.BloodBankAdminID, &c.Document, &c.Status, &c.ContractStart, &c.ContractEnd, &c.CreatedAt, &c.HospitalSignaturePath, &c.AdminSignaturePath, &c.TemplateID)
		if err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}
	return contracts, nil
}

func (r *hospitalRepository) GetHospitalByID(hospitalID string) (*Domain.Hospital, error) {
	query := `SELECT hospital_id, name, address, phone, created_at, latitude, longitude, COALESCE(license_document, '') FROM hospitals WHERE hospital_id = $1`
	hospital := &Domain.Hospital{}
	err := r.db.QueryRow(query, hospitalID).Scan(&hospital.HospitalID, &hospital.Name, &hospital.Address, &hospital.Phone, &hospital.CreatedAt, &hospital.Latitude, &hospital.Longitude, &hospital.LicenseDocument)
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

func (r *hospitalRepository) GetSignedContracts(status string) ([]Domain.HospitalContractResponse, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT c.contract_id, c.hospital_id, h.name as hospital_name, c.blood_bank_admin_id, c.document, 
		                 c.status, c.contract_start, c.contract_end, c.created_at, c.hospital_signature_path, c.admin_signature_path, COALESCE(h.license_document, '')
				  FROM hospital_contracts c
				  JOIN hospitals h ON c.hospital_id = h.hospital_id
				  WHERE c.status = $1
				  ORDER BY c.created_at DESC`
		args = append(args, status)
	} else {
		query = `SELECT c.contract_id, c.hospital_id, h.name as hospital_name, c.blood_bank_admin_id, c.document, 
		                 c.status, c.contract_start, c.contract_end, c.created_at, c.hospital_signature_path, c.admin_signature_path, COALESCE(h.license_document, '')
				  FROM hospital_contracts c
				  JOIN hospitals h ON c.hospital_id = h.hospital_id
				  WHERE c.status IN ($1, $2)
				  ORDER BY c.created_at DESC`
		args = append(args, Domain.ContractStatusApprovedByHospital, Domain.ContractStatusFinalized)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []Domain.HospitalContractResponse
	for rows.Next() {
		var c Domain.HospitalContractResponse
		err := rows.Scan(
			&c.ContractID, &c.HospitalID, &c.HospitalName, &c.BloodBankAdminID, &c.Document,
			&c.Status, &c.ContractStart, &c.ContractEnd, &c.CreatedAt, &c.HospitalSignaturePath, &c.AdminSignaturePath, &c.LicenseDocument,
		)
		if err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}
	return contracts, nil
}

func (r *hospitalRepository) GetHospitalDashboard(hospitalID string) (*Domain.HospitalDashboard, error) {
	dashboard := &Domain.HospitalDashboard{}

	// 1. Get Request counts
	err := r.db.QueryRow(`
		SELECT 
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'FULFILLED'),
			COUNT(*) FILTER (WHERE status = 'APPROVED_PARTIALLY_FULFILLED'),
			COUNT(*) FILTER (WHERE status = 'REJECTED'),
			COUNT(*) FILTER (WHERE status = 'PENDING'),
			COALESCE(SUM(quantity), 0)
		FROM blood_requests 
		WHERE hospital_id = $1`, hospitalID).Scan(
		&dashboard.TotalRequests,
		&dashboard.ApprovedRequests,
		&dashboard.PartiallyFulfilled,
		&dashboard.RejectedRequests,
		&dashboard.PendingRequests,
		&dashboard.TotalUnitsRequested,
	)
	if err != nil {
		return nil, err
	}

	// 2. Get Most Requested Blood Type
	var mostType sql.NullString
	_ = r.db.QueryRow(`
		SELECT blood_type 
		FROM blood_requests 
		WHERE hospital_id = $1 
		GROUP BY blood_type 
		ORDER BY COUNT(*) DESC 
		LIMIT 1`, hospitalID).Scan(&mostType)
	dashboard.MostRequestedBloodType = mostType.String

	// 3. Get Contract Info
	_ = r.db.QueryRow(`
		SELECT status, contract_end 
		FROM hospital_contracts 
		WHERE hospital_id = $1 
		ORDER BY created_at DESC 
		LIMIT 1`, hospitalID).Scan(&dashboard.ContractStatus, &dashboard.ContractEndDate)

	// 4. Get Recent Requests
	rows, err := r.db.Query(`
		SELECT br.request_id, br.hospital_id, h.name, br.blood_type, br.quantity, br.urgency_level, br.status, br.created_at, br.approved_at 
		FROM blood_requests br
		JOIN hospitals h ON br.hospital_id = h.hospital_id
		WHERE br.hospital_id = $1 
		ORDER BY br.created_at DESC 
		LIMIT 5`, hospitalID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var req Domain.BloodRequestResponse
			_ = rows.Scan(&req.RequestID, &req.HospitalID, &req.HospitalName, &req.BloodType, &req.Quantity, &req.UrgencyLevel, &req.Status, &req.CreatedAt, &req.ApprovedAt)
			dashboard.RecentRequests = append(dashboard.RecentRequests, req)
		}
	}

	// 5. Monthly Trends (Last 6 months)
	trendRows, err := r.db.Query(`
		SELECT TO_CHAR(created_at, 'Mon'), COUNT(*) 
		FROM blood_requests 
		WHERE hospital_id = $1 AND created_at > NOW() - INTERVAL '6 months'
		GROUP BY TO_CHAR(created_at, 'Mon'), DATE_TRUNC('month', created_at)
		ORDER BY DATE_TRUNC('month', created_at)`)
	if err == nil {
		defer trendRows.Close()
		for trendRows.Next() {
			var trend Domain.MonthlyTrend
			_ = trendRows.Scan(&trend.Month, &trend.Count)
			dashboard.MonthlyRequestTrends = append(dashboard.MonthlyRequestTrends, trend)
		}
	}

	return dashboard, nil
}

func (r *hospitalRepository) GetHospitalByPhone(phone string) (*Domain.Hospital, error) {
	query := `SELECT hospital_id, name, address, phone, created_at FROM hospitals WHERE phone = $1`
	hospital := &Domain.Hospital{}
	err := r.db.QueryRow(query, phone).Scan(&hospital.HospitalID, &hospital.Name, &hospital.Address, &hospital.Phone, &hospital.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return hospital, nil
}

func (r *hospitalRepository) GetAllHospitals() ([]Domain.Hospital, error) {
	query := `SELECT hospital_id, name, address, phone, COALESCE(license_document, ''), created_at FROM hospitals`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hospitals []Domain.Hospital
	for rows.Next() {
		var h Domain.Hospital
		err := rows.Scan(&h.HospitalID, &h.Name, &h.Address, &h.Phone, &h.LicenseDocument, &h.CreatedAt)
		if err != nil {
			return nil, err
		}
		hospitals = append(hospitals, h)
	}
	return hospitals, nil
}

func (r *hospitalRepository) IsPhoneRegisteredOrPending(phone string) (bool, error) {
	var exists bool

	// 1. Check users table
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1)`, phone).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	// 2. Check hospitals table
	err = r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM hospitals WHERE phone = $1)`, phone).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	// 3. Check pending hospital requests
	err = r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM hospital_requests WHERE phone = $1 AND status = 'PENDING')`, phone).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	// 4. Check pending hospital request admins
	err = r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM hospital_request_admins WHERE admin_phone = $1 AND request_id IN (SELECT request_id FROM hospital_requests WHERE status = 'PENDING'))`, phone).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *hospitalRepository) IsAdminEmailPending(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM hospital_request_admins a
		JOIN hospital_requests r ON a.request_id = r.request_id
		WHERE a.admin_email = $1 AND r.status = 'PENDING'
	)`
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
