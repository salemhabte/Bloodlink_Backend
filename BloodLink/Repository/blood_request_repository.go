package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
	"fmt"
)

type bloodRequestRepository struct {
	db *sql.DB
}

func NewBloodRequestRepository(db *sql.DB) Interfaces.IBloodRequestRepository {
	return &bloodRequestRepository{db: db}
}

func (r *bloodRequestRepository) CreateRequest(req *Domain.BloodRequest) error {
	query := `INSERT INTO blood_requests (request_id, hospital_id, blood_type, component, quantity, urgency_level, status, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(query, req.RequestID, req.HospitalID, req.BloodType, req.Component, req.Quantity, req.UrgencyLevel, req.Status, req.CreatedAt)
	return err
}

func (r *bloodRequestRepository) GetRequestsByHospital(filter Domain.BloodRequestFilter) ([]Domain.BloodRequestResponse, error) {
	query := `SELECT br.request_id, br.hospital_id, h.name,
	                 br.blood_type, br.component, br.quantity, br.urgency_level, br.status,
	                 COALESCE(br.fulfilled_count, 0),
	                 COALESCE(br.fulfilled_quantity_ml, 0),
	                 COALESCE(br.notes, ''),
	                 EXISTS (
	                     SELECT 1
	                     FROM blood_units bu
	                     WHERE bu.request_id = br.request_id
	                       AND bu.status = 'USED'
	                       AND bu.is_deleted = false
	                 ) AS is_used,
	                 br.created_at, br.approved_at
	          FROM blood_requests br
	          JOIN hospitals h ON br.hospital_id = h.hospital_id
	          WHERE br.hospital_id = $1`

	args := []interface{}{filter.HospitalID}
	placeholderID := 2

	if filter.BloodType != "" {
		query += fmt.Sprintf(" AND br.blood_type = $%d", placeholderID)
		args = append(args, filter.BloodType)
		placeholderID++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND br.status = $%d", placeholderID)
		args = append(args, filter.Status)
		placeholderID++
	}
	if filter.UrgencyLevel != "" {
		query += fmt.Sprintf(" AND br.urgency_level = $%d", placeholderID)
		args = append(args, filter.UrgencyLevel)
		placeholderID++
	}
	if filter.Component != "" {
		query += fmt.Sprintf(" AND br.component = $%d", placeholderID)
		args = append(args, filter.Component)
		placeholderID++
	}
	if filter.Component != "" {
		query += fmt.Sprintf(" AND br.component = $%d", placeholderID)
		args = append(args, filter.Component)
		placeholderID++
	}
	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND br.created_at >= $%d", placeholderID)
		args = append(args, filter.StartDate)
		placeholderID++
	}
	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND br.created_at <= $%d", placeholderID)
		args = append(args, filter.EndDate)
		placeholderID++
	}
	query += " ORDER BY CASE WHEN br.urgency_level = 'emergency' THEN 1 ELSE 2 END ASC, br.created_at ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Domain.BloodRequestResponse
	for rows.Next() {
		var req Domain.BloodRequestResponse
		if err := rows.Scan(
			&req.RequestID, &req.HospitalID, &req.HospitalName,
			&req.BloodType, &req.Component, &req.Quantity, &req.UrgencyLevel, &req.Status,
			&req.FulfilledCount, &req.FulfilledQuantityMl, &req.Notes,
			&req.IsUsed,
			&req.CreatedAt, &req.ApprovedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *bloodRequestRepository) GetAllRequests(filter Domain.BloodRequestFilter) ([]Domain.BloodRequestResponse, error) {
	query := `SELECT br.request_id, br.hospital_id, h.name,
	                 br.blood_type, br.component, br.quantity, br.urgency_level, br.status,
	                 COALESCE(br.fulfilled_count, 0),
	                 COALESCE(br.fulfilled_quantity_ml, 0),
	                 COALESCE(br.notes, ''),
	                 EXISTS (
	                     SELECT 1
	                     FROM blood_units bu
	                     WHERE bu.request_id = br.request_id
	                       AND bu.status = 'USED'
	                       AND bu.is_deleted = false
	                 ) AS is_used,
	                 br.created_at, br.approved_at
	          FROM blood_requests br
	          JOIN hospitals h ON br.hospital_id = h.hospital_id
	          WHERE 1=1`

	var args []interface{}
	placeholderID := 1

	if filter.BloodType != "" {
		query += fmt.Sprintf(" AND br.blood_type = $%d", placeholderID)
		args = append(args, filter.BloodType)
		placeholderID++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND br.status = $%d", placeholderID)
		args = append(args, filter.Status)
		placeholderID++
	}
	if filter.UrgencyLevel != "" {
		query += fmt.Sprintf(" AND br.urgency_level = $%d", placeholderID)
		args = append(args, filter.UrgencyLevel)
		placeholderID++
	}
	if filter.Component != "" {
		query += fmt.Sprintf(" AND br.component = $%d", placeholderID)
		args = append(args, filter.Component)
		placeholderID++
	}

	if filter.HospitalID != "" {
		query += fmt.Sprintf(" AND br.hospital_id = $%d", placeholderID)
		args = append(args, filter.HospitalID)
		placeholderID++
	}
	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND br.created_at >= $%d", placeholderID)
		args = append(args, filter.StartDate)
		placeholderID++
	}
	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND br.created_at <= $%d", placeholderID)
		args = append(args, filter.EndDate)
		placeholderID++
	}
	query += " ORDER BY CASE WHEN br.urgency_level = 'emergency' THEN 1 ELSE 2 END ASC, br.created_at ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Domain.BloodRequestResponse
	for rows.Next() {
		var req Domain.BloodRequestResponse
		if err := rows.Scan(
			&req.RequestID, &req.HospitalID, &req.HospitalName,
			&req.BloodType, &req.Component, &req.Quantity, &req.UrgencyLevel, &req.Status,
			&req.FulfilledCount, &req.FulfilledQuantityMl, &req.Notes,
			&req.IsUsed,
			&req.CreatedAt, &req.ApprovedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *bloodRequestRepository) GetRequestByID(requestID string) (*Domain.BloodRequest, error) {
	query := `SELECT request_id, hospital_id, blood_type, component, quantity, urgency_level, status, created_at, approved_at
	          FROM blood_requests WHERE request_id = $1`
	req := &Domain.BloodRequest{}
	err := r.db.QueryRow(query, requestID).Scan(
		&req.RequestID, &req.HospitalID, &req.BloodType, &req.Component, &req.Quantity,
		&req.UrgencyLevel, &req.Status, &req.CreatedAt, &req.ApprovedAt,
	)
	return req, err
}

// UpdateRequestStatus updates status and approved_at (legacy / simple path)
func (r *bloodRequestRepository) UpdateRequestStatus(requestID string, status string, approvedAt *string) error {
	query := `UPDATE blood_requests SET status = $1, approved_at = COALESCE($2, approved_at) WHERE request_id = $3`
	_, err := r.db.Exec(query, status, approvedAt, requestID)
	return err
}

// UpdateRequestStatusWithDetails updates status plus fulfillment summary fields
func (r *bloodRequestRepository) UpdateRequestStatusWithDetails(
	requestID string, status string, approvedAt *string,
	notes string, fulfilledCount int, fulfilledVolumeMl int,
) error {
	query := `
	UPDATE blood_requests
	SET status             = $1,
	    approved_at        = COALESCE($2, approved_at),
	    notes              = $3,
	    fulfilled_count    = $4,
	    fulfilled_quantity_ml = $5
	WHERE request_id = $6`
	_, err := r.db.Exec(query, status, approvedAt, notes, fulfilledCount, fulfilledVolumeMl, requestID)
	return err
}

// GetExpiredReservationRequests returns requests that were approved (FULFILLED/PARTIALLY)
// but whose reserved_at time has passed the cutoff (for auto-rejection by background job)
func (r *bloodRequestRepository) GetExpiredReservationRequests(cutoff string) ([]Domain.BloodRequest, error) {
	query := `
	SELECT request_id, hospital_id, blood_type, component, quantity, urgency_level, status, created_at, approved_at
	FROM blood_requests
	WHERE status IN ('FULFILLED', 'APPROVED_PARTIALLY_FULFILLED')
	  AND approved_at < $1
	`
	rows, err := r.db.Query(query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []Domain.BloodRequest
	for rows.Next() {
		var req Domain.BloodRequest
		if err := rows.Scan(
			&req.RequestID, &req.HospitalID, &req.BloodType, &req.Component, &req.Quantity,
			&req.UrgencyLevel, &req.Status, &req.CreatedAt, &req.ApprovedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}
