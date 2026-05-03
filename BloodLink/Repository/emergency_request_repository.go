package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
	"time"
)

type emergencyRequestRepository struct {
	db *sql.DB
}

func NewEmergencyRequestRepository(db *sql.DB) Interfaces.IEmergencyRequestRepository {
	return &emergencyRequestRepository{db: db}
}

func (r *emergencyRequestRepository) Create(req *Domain.EmergencyRequest) error {
	query := `INSERT INTO emergency_requests (
				emergency_id, request_id, blood_type, quantity_required, 
				quantity_fulfilled, urgency_level, hospital_name, location,
				status, is_manual, created_at, updated_at
			  ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.Exec(query, 
		req.EmergencyID, req.RequestID, req.BloodType, req.QuantityRequired, 
		req.QuantityFulfilled, req.UrgencyLevel, req.HospitalName, req.Location,
		req.Status, req.IsManual, req.CreatedAt, req.UpdatedAt,
	)
	return err
}

func (r *emergencyRequestRepository) UpdateStatus(id string, status string) error {
	var query string
	var err error
	if status == Domain.EmergencyStatusPublished {
		query = `UPDATE emergency_requests SET status = $1, updated_at = $2, published_at = $3 WHERE emergency_id = $4`
		_, err = r.db.Exec(query, status, time.Now(), time.Now(), id)
	} else {
		query = `UPDATE emergency_requests SET status = $1, updated_at = $2 WHERE emergency_id = $3`
		_, err = r.db.Exec(query, status, time.Now(), id)
	}
	return err
}

func (r *emergencyRequestRepository) GetByID(id string) (*Domain.EmergencyRequest, error) {
	query := `SELECT 
				emergency_id, request_id, blood_type, quantity_required, 
				quantity_fulfilled, urgency_level, hospital_name, location,
				status, is_manual, created_at, updated_at, published_at
			  FROM emergency_requests WHERE emergency_id = $1`
	
	req := &Domain.EmergencyRequest{}
	err := r.db.QueryRow(query, id).Scan(
		&req.EmergencyID, &req.RequestID, &req.BloodType, &req.QuantityRequired, 
		&req.QuantityFulfilled, &req.UrgencyLevel, &req.HospitalName, &req.Location,
		&req.Status, &req.IsManual, &req.CreatedAt, &req.UpdatedAt, &req.PublishedAt,
	)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (r *emergencyRequestRepository) GetAll() ([]Domain.EmergencyRequest, error) {
	query := `SELECT 
				emergency_id, request_id, blood_type, quantity_required, 
				quantity_fulfilled, urgency_level, hospital_name, location,
				status, is_manual, created_at, updated_at, published_at
			  FROM emergency_requests ORDER BY created_at DESC`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Domain.EmergencyRequest
	for rows.Next() {
		var req Domain.EmergencyRequest
		err := rows.Scan(
			&req.EmergencyID, &req.RequestID, &req.BloodType, &req.QuantityRequired, 
			&req.QuantityFulfilled, &req.UrgencyLevel, &req.HospitalName, &req.Location,
			&req.Status, &req.IsManual, &req.CreatedAt, &req.UpdatedAt, &req.PublishedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *emergencyRequestRepository) GetActive() ([]Domain.EmergencyRequest, error) {
	query := `SELECT 
				emergency_id, request_id, blood_type, quantity_required, 
				quantity_fulfilled, urgency_level, hospital_name, location,
				status, is_manual, created_at, updated_at, published_at
			  FROM emergency_requests 
			  WHERE status = $1 ORDER BY published_at DESC`
	
	rows, err := r.db.Query(query, Domain.EmergencyStatusPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Domain.EmergencyRequest
	for rows.Next() {
		var req Domain.EmergencyRequest
		err := rows.Scan(
			&req.EmergencyID, &req.RequestID, &req.BloodType, &req.QuantityRequired, 
			&req.QuantityFulfilled, &req.UrgencyLevel, &req.HospitalName, &req.Location,
			&req.Status, &req.IsManual, &req.CreatedAt, &req.UpdatedAt, &req.PublishedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *emergencyRequestRepository) GetByRequestID(requestID string) (*Domain.EmergencyRequest, error) {
	query := `SELECT 
				emergency_id, request_id, blood_type, quantity_required, 
				quantity_fulfilled, urgency_level, hospital_name, location,
				status, is_manual, created_at, updated_at, published_at
			  FROM emergency_requests WHERE request_id = $1`
	
	req := &Domain.EmergencyRequest{}
	err := r.db.QueryRow(query, requestID).Scan(
		&req.EmergencyID, &req.RequestID, &req.BloodType, &req.QuantityRequired, 
		&req.QuantityFulfilled, &req.UrgencyLevel, &req.HospitalName, &req.Location,
		&req.Status, &req.IsManual, &req.CreatedAt, &req.UpdatedAt, &req.PublishedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (r *emergencyRequestRepository) GetByLocation(location string) ([]Domain.EmergencyRequest, error) {
	query := `SELECT 
				emergency_id, request_id, blood_type, quantity_required, 
				quantity_fulfilled, urgency_level, hospital_name, location,
				status, is_manual, created_at, updated_at, published_at
			  FROM emergency_requests 
			  WHERE status = $1 AND location ILIKE $2
			  ORDER BY published_at DESC`
	
	rows, err := r.db.Query(query, Domain.EmergencyStatusPublished, "%"+location+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Domain.EmergencyRequest
	for rows.Next() {
		var req Domain.EmergencyRequest
		err := rows.Scan(
			&req.EmergencyID, &req.RequestID, &req.BloodType, &req.QuantityRequired, 
			&req.QuantityFulfilled, &req.UrgencyLevel, &req.HospitalName, &req.Location,
			&req.Status, &req.IsManual, &req.CreatedAt, &req.UpdatedAt, &req.PublishedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}
