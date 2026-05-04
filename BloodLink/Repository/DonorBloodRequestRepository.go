package Repository

import (
	"bloodlink/Domain"
	"database/sql"
	"errors"
)

type DonorBloodRequestRepository struct {
	db *sql.DB
}

func NewDonorBloodRequestRepository(db *sql.DB) *DonorBloodRequestRepository {
	return &DonorBloodRequestRepository{db: db}
}

///////////////////////
// CREATE REQUEST
///////////////////////

func (r *DonorBloodRequestRepository) Create(req *Domain.DonorBloodRequest) error {
	query := `
	INSERT INTO donor_blood_requests
	(request_id, donor_id, blood_type, quantity_ml, reason, priority_score, status, created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`

	_, err := r.db.Exec(query,
		req.RequestID,
		req.DonorID,
		req.BloodType,
		req.QuantityML,
		req.Reason,
		req.PriorityScore,
		req.Status,
		req.CreatedAt,
	)
	return err
}

///////////////////////
// GET ALL REQUESTS
///////////////////////

func (r *DonorBloodRequestRepository) GetAll() ([]Domain.DonorBloodRequest, error) {
	query := `
	SELECT request_id, donor_id, blood_type, quantity_ml, reason, priority_score, status, created_at
	FROM donor_blood_requests
	ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Domain.DonorBloodRequest

	for rows.Next() {
		var req Domain.DonorBloodRequest

		err := rows.Scan(
			&req.RequestID,
			&req.DonorID,
			&req.BloodType,
			&req.QuantityML,
			&req.Reason,
			&req.PriorityScore,
			&req.Status,
			&req.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, req)
	}

	return result, nil
}

///////////////////////
// UPDATE STATUS
///////////////////////

func (r *DonorBloodRequestRepository) UpdateStatus(id, status string) error {
	query := `UPDATE donor_blood_requests SET status=$1 WHERE request_id=$2`
	_, err := r.db.Exec(query, status, id)
	return err
}

///////////////////////
// CHECK AVAILABLE BLOOD (ML)
///////////////////////

func (r *DonorBloodRequestRepository) GetAvailableBloodVolume(bloodType string) (int, error) {
	query := `
	SELECT COALESCE(SUM(volume_ml),0)
	FROM blood_units
	WHERE blood_type = $1 AND status = 'AVAILABLE'
	`

	var total int
	err := r.db.QueryRow(query, bloodType).Scan(&total)
	return total, err
}

///////////////////////
// 🔥 PROPER FULFILL (FIFO + ML BASED)
///////////////////////

func (r *DonorBloodRequestRepository) FulfillRequest(bloodType string, requiredML int) error {

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// 1. Get available blood ordered by oldest (FIFO)
	query := `
	SELECT blood_unit_id, volume_ml
	FROM blood_units
	WHERE blood_type = $1 AND status = 'AVAILABLE'
	ORDER BY collection_date ASC
	`

	rows, err := tx.Query(query, bloodType)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer rows.Close()

	totalUsed := 0

	for rows.Next() {
		var unitID string
		var volume int

		if err := rows.Scan(&unitID, &volume); err != nil {
			tx.Rollback()
			return err
		}

		// mark this unit as USED
		_, err := tx.Exec(`
			UPDATE blood_units
			SET status = 'USED'
			WHERE blood_unit_id = $1
		`, unitID)

		if err != nil {
			tx.Rollback()
			return err
		}

		totalUsed += volume

		// stop when enough blood collected
		if totalUsed >= requiredML {
			break
		}
	}

	//  not enough blood
	if totalUsed < requiredML {
		tx.Rollback()
		return errors.New("not enough blood available")
	}

	return tx.Commit()
}
// GET BY ID
func (r *DonorBloodRequestRepository) GetByID(id string) (*Domain.DonorBloodRequest, error) {
	query := `
	SELECT request_id, donor_id, blood_type, quantity_ml, reason, priority_score, status, created_at
	FROM donor_blood_requests
	WHERE request_id = $1
	`

	var req Domain.DonorBloodRequest

	err := r.db.QueryRow(query, id).Scan(
		&req.RequestID,
		&req.DonorID,
		&req.BloodType,
		&req.QuantityML,
		&req.Reason,
		&req.PriorityScore,
		&req.Status,
		&req.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &req, nil
}
// COUNT DONATIONS
func (r *DonorBloodRequestRepository) CountDonationsByDonor(donorID string) (int, error) {
	query := `
	SELECT COUNT(*)
	FROM donor_test_results
	WHERE donor_id = $1 AND overall_status = 'CLEARED'
	`

	var count int
	err := r.db.QueryRow(query, donorID).Scan(&count)
	return count, err
}
// GET BLOOD TYPE FROM SYSTEM
func (r *DonorBloodRequestRepository) GetDonorBloodType(donorID string) (string, error) {
	query := `
	SELECT blood_type
	FROM donor_test_results
	WHERE donor_id = $1 AND overall_status = 'CLEARED'
	LIMIT 1
	`

	var bloodType string
	err := r.db.QueryRow(query, donorID).Scan(&bloodType)
	return bloodType, err
}
func (r *DonorBloodRequestRepository) GetByDonorID(donorID string) ([]Domain.DonorBloodRequest, error) {
	query := `
	SELECT request_id, donor_id, blood_type, quantity_ml, reason, priority_score, status, created_at
	FROM donor_blood_requests
	WHERE donor_id = $1
	ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, donorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Domain.DonorBloodRequest

	for rows.Next() {
		var rqs Domain.DonorBloodRequest
		err := rows.Scan(
			&rqs.RequestID,
			&rqs.DonorID,
			&rqs.BloodType,
			&rqs.QuantityML,
			&rqs.Reason,
			&rqs.PriorityScore,
			&rqs.Status,
			&rqs.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, rqs)
	}

	return result, nil
}
func (r *DonorBloodRequestRepository) ReserveBlood(bloodType string, quantity int) error {
	query := `
	UPDATE blood_units
	SET status = 'RESERVED'
	WHERE blood_unit_id IN (
		SELECT blood_unit_id FROM blood_units
		WHERE blood_type = $1 AND status = 'AVAILABLE'
		ORDER BY expiration_date ASC
		LIMIT $2
	)
	`
	_, err := r.db.Exec(query, bloodType, quantity)
	return err
}
func (r *DonorBloodRequestRepository) MarkReservedAsUsed(bloodType string, quantity int) error {
	query := `
	UPDATE blood_units
	SET status = 'USED'
	WHERE blood_unit_id IN (
		SELECT blood_unit_id FROM blood_units
		WHERE blood_type = $1 AND status = 'RESERVED'
		ORDER BY expiration_date ASC
		LIMIT $2
	)
	`
	_, err := r.db.Exec(query, bloodType, quantity)
	return err
}