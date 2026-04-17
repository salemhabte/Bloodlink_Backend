package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
)

type bloodRequestRepository struct {
	db *sql.DB
}

func NewBloodRequestRepository(db *sql.DB) Interfaces.IBloodRequestRepository {
	return &bloodRequestRepository{db: db}
}

func (r *bloodRequestRepository) CreateRequest(req *Domain.BloodRequest) error {
	query := `INSERT INTO blood_requests (request_id, hospital_id, blood_type, quantity, urgency_level, status, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, req.RequestID, req.HospitalID, req.BloodType, req.Quantity, req.UrgencyLevel, req.Status, req.CreatedAt)
	return err
}

func (r *bloodRequestRepository) GetRequestsByHospital(hospitalID string) ([]Domain.BloodRequestResponse, error) {
	query := `SELECT br.request_id, br.hospital_id, h.name as hospital_name, br.blood_type, br.quantity, br.urgency_level, br.status, br.created_at, br.approved_at 
			  FROM blood_requests br
			  JOIN hospitals h ON br.hospital_id = h.hospital_id
			  WHERE br.hospital_id = $1 ORDER BY br.created_at DESC`

	rows, err := r.db.Query(query, hospitalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Domain.BloodRequestResponse
	for rows.Next() {
		var req Domain.BloodRequestResponse
		if err := rows.Scan(&req.RequestID, &req.HospitalID, &req.HospitalName, &req.BloodType, &req.Quantity, &req.UrgencyLevel, &req.Status, &req.CreatedAt, &req.ApprovedAt); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *bloodRequestRepository) GetAllRequests() ([]Domain.BloodRequestResponse, error) {
	query := `SELECT br.request_id, br.hospital_id, h.name as hospital_name, br.blood_type, br.quantity, br.urgency_level, br.status, br.created_at, br.approved_at 
			  FROM blood_requests br
			  JOIN hospitals h ON br.hospital_id = h.hospital_id
			  ORDER BY br.created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Domain.BloodRequestResponse
	for rows.Next() {
		var req Domain.BloodRequestResponse
		if err := rows.Scan(&req.RequestID, &req.HospitalID, &req.HospitalName, &req.BloodType, &req.Quantity, &req.UrgencyLevel, &req.Status, &req.CreatedAt, &req.ApprovedAt); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *bloodRequestRepository) GetRequestByID(requestID string) (*Domain.BloodRequest, error) {
	query := `SELECT request_id, hospital_id, blood_type, quantity, urgency_level, status, created_at, approved_at 
			  FROM blood_requests WHERE request_id = $1`
	req := &Domain.BloodRequest{}
	err := r.db.QueryRow(query, requestID).Scan(&req.RequestID, &req.HospitalID, &req.BloodType, &req.Quantity, &req.UrgencyLevel, &req.Status, &req.CreatedAt, &req.ApprovedAt)
	return req, err
}

func (r *bloodRequestRepository) UpdateRequestStatus(requestID string, status string, approvedAt *string) error {
	query := `UPDATE blood_requests SET status = $1, approved_at = COALESCE($2, approved_at) WHERE request_id = $3`
	_, err := r.db.Exec(query, status, approvedAt, requestID)
	return err
}
