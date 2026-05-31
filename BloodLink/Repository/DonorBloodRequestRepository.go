package Repository

import (
	domain "bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
	"fmt"
	"time"
)

// struct
type donorBloodRequestRepository struct {
	db *sql.DB
}

// compile-time check (IMPORTANT)
var _ Interfaces.IDonorBloodRequestRepository = &donorBloodRequestRepository{}

// constructor
func NewDonorBloodRequestRepository(db *sql.DB) Interfaces.IDonorBloodRequestRepository {
	return &donorBloodRequestRepository{db: db}
}
func (r *donorBloodRequestRepository) Create(req *domain.DonorBloodRequest) error {

	query := `
	INSERT INTO donor_blood_requests (
		request_id, donor_id,
		donor_name, donor_email, donor_phone, donor_address,
		blood_type, component_type, units, reason,
		hospital_name, hospital_address, hospital_phone,
		status, created_at
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`

	_, err := r.db.Exec(query,
		req.RequestID,
		req.DonorID,

		req.DonorName,
		req.DonorEmail,
		req.DonorPhone,
		req.DonorAddress,

		req.BloodType,
		req.ComponentType,
		req.Units,
		req.Reason,

		req.HospitalName,
		req.HospitalAddress,
		req.HospitalPhone,

		req.Status,
		req.CreatedAt,
	)

	return err
}
func (r *donorBloodRequestRepository) GetAll() ([]domain.DonorBloodRequest, error) {

	query := `
	SELECT
		request_id, donor_id,
		donor_name, donor_email, donor_phone, donor_address,
		blood_type, COALESCE(component_type, ''), units, reason,
		hospital_name, hospital_address, hospital_phone,
		status, created_at, COALESCE(reserved_units, 0)
	FROM donor_blood_requests
	ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []domain.DonorBloodRequest{}

	for rows.Next() {
		var rqs domain.DonorBloodRequest

		err := rows.Scan(
			&rqs.RequestID,
			&rqs.DonorID,

			&rqs.DonorName,
			&rqs.DonorEmail,
			&rqs.DonorPhone,
			&rqs.DonorAddress,

			&rqs.BloodType,
			&rqs.ComponentType,
			&rqs.Units,
			&rqs.Reason,

			&rqs.HospitalName,
			&rqs.HospitalAddress,
			&rqs.HospitalPhone,

		
			&rqs.Status,
			&rqs.CreatedAt,
			&rqs.ReservedUnits,
		)

		if err != nil {
			return nil, err
		}

		rqs.CanFulfill = (rqs.Status == "APPROVED" || rqs.Status == "PARTIALLY APPROVED")
		result = append(result, rqs)
	}

	return result, nil
}
func (r *donorBloodRequestRepository) GetByID(id string) (*domain.DonorBloodRequest, error) {

	query := `
	SELECT
		request_id, donor_id,
		donor_name, donor_email, donor_phone, donor_address,
		blood_type, COALESCE(component_type, ''), units, reason,
		hospital_name, hospital_address, hospital_phone,
		status, created_at, COALESCE(reserved_units, 0)
	FROM donor_blood_requests
	WHERE request_id=$1
	`

	var rqs domain.DonorBloodRequest

	err := r.db.QueryRow(query, id).Scan(
		&rqs.RequestID,
		&rqs.DonorID,

		&rqs.DonorName,
		&rqs.DonorEmail,
		&rqs.DonorPhone,
		&rqs.DonorAddress,

		&rqs.BloodType,
		&rqs.ComponentType,
		&rqs.Units,
		&rqs.Reason,

		&rqs.HospitalName,
		&rqs.HospitalAddress,
		&rqs.HospitalPhone,

		&rqs.Status,
		&rqs.CreatedAt,
		&rqs.ReservedUnits,
	)

	if err != nil {
		return nil, err
	}

	rqs.CanFulfill = (rqs.Status == "APPROVED" || rqs.Status == "PARTIALLY APPROVED")
	return &rqs, nil
}
func (r *donorBloodRequestRepository) GetByDonorID(donorID string, filter domain.DonorBloodRequestFilter) ([]domain.DonorBloodRequest, error) {

	query := `
	SELECT
		request_id, donor_id,
		donor_name, donor_email, donor_phone, donor_address,
		blood_type, COALESCE(component_type, ''), units, reason,
		hospital_name, hospital_address, hospital_phone,
		status, created_at, COALESCE(reserved_units, 0)
	FROM donor_blood_requests
	WHERE donor_id=$1
	`
	args := []interface{}{donorID}
	argId := 2

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status=$%d", argId)
		args = append(args, filter.Status)
		argId++
	}
	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND created_at >= $%d", argId)
		args = append(args, filter.StartDate)
		argId++
	}
	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND created_at <= $%d", argId)
		args = append(args, filter.EndDate)
		argId++
	}

	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize as empty slice to avoid 'null' in JSON
	result := []domain.DonorBloodRequest{}

	for rows.Next() {
		var rqs domain.DonorBloodRequest

		err := rows.Scan(
			&rqs.RequestID,
			&rqs.DonorID,

			&rqs.DonorName,
			&rqs.DonorEmail,
			&rqs.DonorPhone,
			&rqs.DonorAddress,

			&rqs.BloodType,
			&rqs.ComponentType,
			&rqs.Units,
			&rqs.Reason,

			&rqs.HospitalName,
			&rqs.HospitalAddress,
			&rqs.HospitalPhone,

			&rqs.Status,
			&rqs.CreatedAt,
			&rqs.ReservedUnits,
		)

		if err != nil {
			return nil, err
		}

		rqs.CanFulfill = (rqs.Status == "APPROVED" || rqs.Status == "PARTIALLY APPROVED")
		result = append(result, rqs)
	}

	return result, nil
}
func (r *donorBloodRequestRepository) UpdateStatus(id, status string) error {
	_, err := r.db.Exec(`
	UPDATE donor_blood_requests
	SET status=$1
	WHERE request_id=$2
	`, status, id)
	return err
}

func (r *donorBloodRequestRepository) UpdateStatusWithUnits(id, status string, reservedUnits int) error {
	_, err := r.db.Exec(`
	UPDATE donor_blood_requests
	SET status=$1, reserved_units=$2
	WHERE request_id=$3
	`, status, reservedUnits, id)
	return err
}
func (r *donorBloodRequestRepository) GetAvailableBloodUnits(bloodType string) ([]string, error) {

	query := `
	SELECT blood_unit_id
	FROM blood_units
	WHERE blood_type=$1
	AND status='AVAILABLE'
	AND expiration_date > NOW()
	ORDER BY expiration_date ASC
	`

	rows, err := r.db.Query(query, bloodType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}
func (r *donorBloodRequestRepository) ReserveBloodUnits(
	requestID string,
	bloodType string,
	componentType string,
	requiredUnits int,
) (int, error) {

	query := `
	SELECT blood_unit_id
	FROM blood_units
	WHERE blood_type=$1 AND component_type=$2
	AND status='AVAILABLE'
	AND expiration_date > NOW()
	ORDER BY expiration_date ASC
	LIMIT $3
	`

	rows, err := r.db.Query(query, bloodType, componentType, requiredUnits)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var selected []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		selected = append(selected, id)
	}

	// If we don't have ANY units
	if len(selected) == 0 {
		return 0, nil
	}

	// ✅ RESERVE
	for _, id := range selected {
		_, err := r.db.Exec(`
			UPDATE blood_units
			SET status='RESERVED',
			    donor_request_id=$1,
				reserved_at=NOW()
			WHERE blood_unit_id=$2
		`, requestID, id)

		if err != nil {
			return 0, err
		}
	}

	return len(selected), nil
}
func (r *donorBloodRequestRepository) MarkReservedAsUsed(requestID string) error {

	_, err := r.db.Exec(`
	UPDATE blood_units
	SET status='USED'
	WHERE donor_request_id=$1
	AND status='RESERVED'
	`, requestID)

	return err
}

func (r *donorBloodRequestRepository) GetDonorIDByUserID(userID string) (string, error) {
	var donorID string
	err := r.db.QueryRow(`SELECT donor_id FROM donors WHERE user_id=$1`, userID).Scan(&donorID)
	return donorID, err
}

func (r *donorBloodRequestRepository) GetDonorProfile(donorID string) (*domain.DonorProfile, error) {
	query := `
		SELECT 
			u.full_name, 
			u.email, 
			u.phone, 
			COALESCE(p.address, ''), 
			COALESCE(d.blood_type, '')
		FROM donors d
		JOIN users u ON d.user_id = u.user_id
		LEFT JOIN user_profiles p ON u.user_id = p.user_id
		WHERE d.donor_id = $1
	`
	var profile domain.DonorProfile
	err := r.db.QueryRow(query, donorID).Scan(
		&profile.FullName,
		&profile.Email,
		&profile.Phone,
		&profile.Address,
		&profile.BloodType,
	)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *donorBloodRequestRepository) HasSuccessfulDonation(donorID string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM blood_units bu
		JOIN donation_records d ON bu.donation_id = d.donation_id
		WHERE d.donor_id = $1
	`
	var count int
	err := r.db.QueryRow(query, donorID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *donorBloodRequestRepository) GetAllAdmin(filter domain.DonorBloodRequestFilter) ([]domain.DonorBloodRequest, error) {
	query := `
	SELECT
		dbr.request_id, dbr.donor_id,
		dbr.donor_name, dbr.donor_email, dbr.donor_phone, dbr.donor_address,
		dbr.blood_type, COALESCE(dbr.component_type, ''), dbr.units, dbr.reason,
		dbr.hospital_name, dbr.hospital_address, dbr.hospital_phone,
		dbr.status, dbr.created_at, COALESCE(dbr.reserved_units, 0),
		(
			SELECT COUNT(*)
			FROM blood_units bu
			JOIN donation_records d ON bu.donation_id = d.donation_id
			WHERE d.donor_id = dbr.donor_id
		) as successful_donations
	FROM donor_blood_requests dbr
	WHERE 1=1
	`
	var args []interface{}
	argId := 1

	if filter.BloodType != "" {
		query += fmt.Sprintf(" AND dbr.blood_type=$%d", argId)
		args = append(args, filter.BloodType)
		argId++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND dbr.status=$%d", argId)
		args = append(args, filter.Status)
		argId++
	}
	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND dbr.created_at >= $%d", argId)
		args = append(args, filter.StartDate)
		argId++
	}
	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND dbr.created_at <= $%d", argId)
		args = append(args, filter.EndDate)
		argId++
	}

	query += ` ORDER BY 
		CASE WHEN dbr.status = 'PENDING' THEN 0 ELSE 1 END ASC,
		successful_donations DESC, 
		dbr.created_at DESC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []domain.DonorBloodRequest{}

	for rows.Next() {
		var rqs domain.DonorBloodRequest
		err := rows.Scan(
			&rqs.RequestID,
			&rqs.DonorID,
			&rqs.DonorName,
			&rqs.DonorEmail,
			&rqs.DonorPhone,
			&rqs.DonorAddress,
			&rqs.BloodType,
			&rqs.ComponentType,
			&rqs.Units,
			&rqs.Reason,
			&rqs.HospitalName,
			&rqs.HospitalAddress,
			&rqs.HospitalPhone,
			&rqs.Status,
			&rqs.CreatedAt,
			&rqs.ReservedUnits,
			&rqs.SuccessfulDonations,
		)
		if err != nil {
			return nil, err
		}
		rqs.CanFulfill = (rqs.Status == "APPROVED" || rqs.Status == "PARTIALLY APPROVED")
		result = append(result, rqs)
	}
	return result, nil
}

func (r *donorBloodRequestRepository) ExpireStaleReservations() error {
	// 1. Find all blood units that have been reserved for more than 24 hours
	// 2. Change their status to AVAILABLE
	// 3. Mark the corresponding donor_blood_requests as REJECTED

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := `
		UPDATE donor_blood_requests
		SET status = 'REJECTED'
		WHERE request_id IN (
			SELECT donor_request_id
			FROM blood_units
			WHERE status = 'RESERVED'
			  AND donor_request_id IS NOT NULL
			  AND reserved_at < NOW() - INTERVAL '24 hours'
		)
	`
	_, err = tx.Exec(query)
	if err != nil {
		tx.Rollback()
		return err
	}

	query2 := `
		UPDATE blood_units
		SET status = 'AVAILABLE',
		    donor_request_id = NULL,
			reserved_at = NULL
		WHERE status = 'RESERVED'
		  AND donor_request_id IS NOT NULL
		  AND reserved_at < NOW() - INTERVAL '24 hours'
	`
	_, err = tx.Exec(query2)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// IsDonorInTop10 checks if the donor ranks in the top 10 by successful donations (CLEARED test results).
func (r *donorBloodRequestRepository) IsDonorInTop10(donorID string) (bool, error) {
	query := `
		SELECT donor_id FROM (
			SELECT d.donor_id, COUNT(dtr.test_id) AS donation_count
			FROM donors d
			LEFT JOIN donor_test_results dtr
				ON dtr.donor_id = d.donor_id AND dtr.overall_status = 'CLEARED'
			GROUP BY d.donor_id
			ORDER BY donation_count DESC
			LIMIT 10
		) top10
		WHERE donor_id = $1
	`
	var found string
	err := r.db.QueryRow(query, donorID).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetLastRequestDateByDonor returns the most recent request date for a donor.
// Returns zero time and sql.ErrNoRows if no request exists (safe to ignore in usecase).
func (r *donorBloodRequestRepository) GetLastRequestDateByDonor(donorID string) (time.Time, error) {
	var t time.Time
	err := r.db.QueryRow(`
		SELECT created_at
		FROM donor_blood_requests
		WHERE donor_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, donorID).Scan(&t)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	return t, err
}
