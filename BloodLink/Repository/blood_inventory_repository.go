package Repository

import (
	"bloodlink/Domain"
	"database/sql"
	"fmt"
	"time"
)

type BloodInventoryRepository struct {
	DB *sql.DB
}

func NewBloodInventoryRepository(db *sql.DB) *BloodInventoryRepository {
	return &BloodInventoryRepository{DB: db}
}

// GetAllBloodUnits returns all blood units
func (r *BloodInventoryRepository) GetAllBloodUnits() ([]Domain.BloodUnit, error) {
	query := `
	SELECT blood_unit_id, donation_id, blood_type, COALESCE(component_type,''),
	       volume_ml, collection_date, expiration_date, status,
	       COALESCE(reserved_for_hospital_id,''), reserved_at, COALESCE(request_id,''), created_at
	FROM blood_units
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []Domain.BloodUnit
	for rows.Next() {
		var u Domain.BloodUnit
		err := rows.Scan(
			&u.BloodUnitID, &u.DonationID, &u.BloodType, &u.ComponentType,
			&u.VolumeML, &u.CollectionDate, &u.ExpirationDate, &u.Status,
			&u.ReservedForHospitalID, &u.ReservedAt, &u.RequestID, &u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, nil
}

// GetBloodUnitByID returns a single unit by ID
func (r *BloodInventoryRepository) GetBloodUnitByID(id string) (*Domain.BloodUnit, error) {
	query := `
	SELECT blood_unit_id, donation_id, blood_type, COALESCE(component_type,''),
	       volume_ml, collection_date, expiration_date, status,
	       COALESCE(reserved_for_hospital_id,''), reserved_at, COALESCE(request_id,''), created_at
	FROM blood_units WHERE blood_unit_id = $1
	`
	var u Domain.BloodUnit
	err := r.DB.QueryRow(query, id).Scan(
		&u.BloodUnitID, &u.DonationID, &u.BloodType, &u.ComponentType,
		&u.VolumeML, &u.CollectionDate, &u.ExpirationDate, &u.Status,
		&u.ReservedForHospitalID, &u.ReservedAt, &u.RequestID, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateBloodUnitStatus updates the status of a blood unit
func (r *BloodInventoryRepository) UpdateBloodUnitStatus(id string, status string) error {
	query := `UPDATE blood_units SET status=$1 WHERE blood_unit_id=$2`
	_, err := r.DB.Exec(query, status, id)
	return err
}

// DeleteBloodUnitByID deletes a blood unit (legacy, no audit)
func (r *BloodInventoryRepository) DeleteBloodUnitByID(id string) error {
	query := `DELETE FROM blood_units WHERE blood_unit_id=$1`
	_, err := r.DB.Exec(query, id)
	return err
}

// GetFullBloodUnitDetails returns enriched details for one unit
func (r *BloodInventoryRepository) GetFullBloodUnitDetails(id string) (map[string]interface{}, error) {
	query := `
SELECT
    bu.blood_unit_id, bu.blood_type, bu.volume_ml,
    bu.collection_date, bu.expiration_date, bu.status,
    d.donation_id, d.donor_id, d.collected_by,
    u.full_name, u.email, u.phone
FROM blood_units bu
JOIN donation_records d ON bu.donation_id = d.donation_id
JOIN donors dn ON d.donor_id = dn.donor_id
JOIN users u ON dn.user_id = u.user_id
WHERE bu.blood_unit_id = $1
`
	var result = make(map[string]interface{})
	row := r.DB.QueryRow(query, id)

	var bloodUnit Domain.BloodUnit
	var donationID, donorID, collectedBy string
	var donorName, donorEmail, donorPhone string

	err := row.Scan(
		&bloodUnit.BloodUnitID, &bloodUnit.BloodType, &bloodUnit.VolumeML,
		&bloodUnit.CollectionDate, &bloodUnit.ExpirationDate, &bloodUnit.Status,
		&donationID, &donorID, &collectedBy,
		&donorName, &donorEmail, &donorPhone,
	)
	if err != nil {
		return nil, err
	}

	donation := map[string]interface{}{
		"donation_id":  donationID,
		"donor_id":     donorID,
		"collected_by": collectedBy,
	}
	donor := map[string]interface{}{
		"name":  donorName,
		"email": donorEmail,
		"phone": donorPhone,
	}

	rows, err := r.DB.Query(`
	SELECT hiv_result, hepatitis_result, syphilis_result
	FROM donor_test_results WHERE donation_id = $1
	`, donation["donation_id"])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []map[string]string
	for rows.Next() {
		var hiv, hep, syph string
		rows.Scan(&hiv, &hep, &syph)
		tests = append(tests, map[string]string{
			"hiv": hiv, "hepatitis": hep, "syphilis": syph,
		})
	}

	result["blood_unit"] = bloodUnit
	result["donor"] = donor
	result["donation"] = donation
	result["test_results"] = tests
	return result, nil
}

// FilterBloodUnits filters blood units by various criteria
func (r *BloodInventoryRepository) FilterBloodUnits(
	unitID, bloodType, status string,
	startDate, endDate string,
) ([]Domain.BloodUnit, error) {

	query := `
	SELECT blood_unit_id, donation_id, blood_type, COALESCE(component_type,''),
	       volume_ml, collection_date, expiration_date, status,
	       COALESCE(reserved_for_hospital_id,''), reserved_at, COALESCE(request_id,''), created_at
	FROM blood_units
	WHERE 1=1
	`
	args := []interface{}{}

	if unitID != "" {
		args = append(args, "%"+unitID+"%")
		query += fmt.Sprintf(" AND blood_unit_id LIKE $%d", len(args))
	}
	if bloodType != "" {
		args = append(args, bloodType)
		query += fmt.Sprintf(" AND blood_type = $%d", len(args))
	}
	if status != "" {
		args = append(args, status)
		query += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if startDate != "" && endDate != "" {
		args = append(args, startDate, endDate)
		query += fmt.Sprintf(" AND collection_date BETWEEN $%d AND $%d", len(args)-1, len(args))
	}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []Domain.BloodUnit
	for rows.Next() {
		var u Domain.BloodUnit
		if err := rows.Scan(
			&u.BloodUnitID, &u.DonationID, &u.BloodType, &u.ComponentType,
			&u.VolumeML, &u.CollectionDate, &u.ExpirationDate, &u.Status,
			&u.ReservedForHospitalID, &u.ReservedAt, &u.RequestID, &u.CreatedAt,
		); err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, nil
}

// MarkExpiredUnits marks blood units past expiration date as EXPIRED
func (r *BloodInventoryRepository) MarkExpiredUnits() error {
	query := `
	UPDATE blood_units
	SET status = 'EXPIRED'
	WHERE expiration_date < NOW()
	AND status NOT IN ('EXPIRED', 'USED')
	`
	_, err := r.DB.Exec(query)
	return err
}

// CountAvailableUnitsByBloodType counts available (non-reserved, non-expired) units
func (r *BloodInventoryRepository) CountAvailableUnitsByBloodType(bloodType string) (int, error) {
	query := `SELECT COUNT(*) FROM blood_units WHERE blood_type = $1 AND status = 'AVAILABLE' AND expiration_date > NOW()`
	var count int
	err := r.DB.QueryRow(query, bloodType).Scan(&count)
	return count, err
}

// ConsumeUnits marks units as FULFILLED (legacy, still used by emergency)
func (r *BloodInventoryRepository) ConsumeUnits(bloodType string, quantity int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	UPDATE blood_units
	SET status = 'FULFILLED'
	WHERE blood_unit_id IN (
		SELECT blood_unit_id
		FROM blood_units
		WHERE blood_type = $1 AND status = 'AVAILABLE' AND expiration_date > NOW()
		ORDER BY expiration_date ASC
		LIMIT $2
		FOR UPDATE
	)
	`
	res, err := tx.Exec(query, bloodType, quantity)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if int(rowsAffected) < quantity {
		return fmt.Errorf("insufficient inventory: requested %d, but only %d available", quantity, rowsAffected)
	}
	return tx.Commit()
}

// ReserveUnitsForHospital reserves blood units (FIFO by expiry) for a specific hospital request.
// Returns the list of reserved units (may be less than quantity if partially available).
func (r *BloodInventoryRepository) ReserveUnitsForHospital(bloodType string, quantity int, hospitalID string, requestID string) ([]Domain.BloodUnit, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()

	// Select AVAILABLE units ordered by nearest expiry (FIFO), lock rows
	selectQuery := `
	SELECT blood_unit_id, donation_id, blood_type, COALESCE(component_type,''),
	       volume_ml, collection_date, expiration_date, status, created_at
	FROM blood_units
	WHERE blood_type = $1 AND status = 'AVAILABLE' AND expiration_date > NOW()
	ORDER BY expiration_date ASC
	FOR UPDATE SKIP LOCKED
	`
	rows, err := tx.Query(selectQuery, bloodType)
	if err != nil {
		return nil, err
	}

	var units []Domain.BloodUnit
	accumulatedVolume := 0
	
	for rows.Next() {
		var u Domain.BloodUnit
		if err := rows.Scan(
			&u.BloodUnitID, &u.DonationID, &u.BloodType, &u.ComponentType,
			&u.VolumeML, &u.CollectionDate, &u.ExpirationDate, &u.Status, &u.CreatedAt,
		); err != nil {
			rows.Close()
			return nil, err
		}
		
		units = append(units, u)
		accumulatedVolume += u.VolumeML
		
		if accumulatedVolume >= quantity {
			break // Stop once we have enough volume
		}
	}
	rows.Close()

	if len(units) == 0 {
		return nil, nil // nothing available
	}

	// Reserve each selected unit
	for _, u := range units {
		_, err := tx.Exec(`
			UPDATE blood_units
			SET status = 'RESERVED',
			    reserved_for_hospital_id = $1,
			    reserved_at = $2,
			    request_id = $3
			WHERE blood_unit_id = $4
		`, hospitalID, now, requestID, u.BloodUnitID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Populate reservation fields in returned slice
	for i := range units {
		units[i].Status = "RESERVED"
		units[i].ReservedForHospitalID = hospitalID
		units[i].ReservedAt = &now
		units[i].RequestID = requestID
	}
	return units, nil
}

// MarkUnitAsUsed transitions a RESERVED unit to USED. Fails if not currently RESERVED.
func (r *BloodInventoryRepository) MarkUnitAsUsed(unitID string) error {
	result, err := r.DB.Exec(`
		UPDATE blood_units SET status = 'USED'
		WHERE blood_unit_id = $1 AND status = 'RESERVED'
	`, unitID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("unit not found or not in RESERVED status")
	}
	return nil
}

// ExpireStaleReservations finds units that have been RESERVED for longer than cutoff,
// releases them back to AVAILABLE, and returns the affected blood_request IDs so the
// caller can set those requests to REJECTED.
func (r *BloodInventoryRepository) ExpireStaleReservations(cutoff time.Time) ([]string, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Collect request IDs linked to stale reservations
	rows, err := tx.Query(`
		SELECT DISTINCT request_id
		FROM blood_units
		WHERE status = 'RESERVED' AND reserved_at < $1 AND request_id IS NOT NULL
	`, cutoff)
	if err != nil {
		return nil, err
	}
	var requestIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		requestIDs = append(requestIDs, id)
	}
	rows.Close()

	if len(requestIDs) == 0 {
		return nil, nil
	}

	// Release stale reservations back to AVAILABLE
	_, err = tx.Exec(`
		UPDATE blood_units
		SET status = 'AVAILABLE',
		    reserved_for_hospital_id = NULL,
		    reserved_at = NULL,
		    request_id = NULL
		WHERE status = 'RESERVED' AND reserved_at < $1
	`, cutoff)
	if err != nil {
		return nil, err
	}

	return requestIDs, tx.Commit()
}

// GetReservedUnitsByHospitalID returns all units currently reserved for a hospital.
// Used by admin to see what blood is set aside when a hospital arrives.
func (r *BloodInventoryRepository) GetReservedUnitsByHospitalID(hospitalID string) ([]Domain.BloodUnit, error) {
	query := `
	SELECT blood_unit_id, donation_id, blood_type, COALESCE(component_type,''),
	       volume_ml, collection_date, expiration_date, status,
	       COALESCE(reserved_for_hospital_id,''), reserved_at, COALESCE(request_id,''), created_at
	FROM blood_units
	WHERE status = 'RESERVED' AND reserved_for_hospital_id = $1
	ORDER BY expiration_date ASC
	`
	rows, err := r.DB.Query(query, hospitalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []Domain.BloodUnit
	for rows.Next() {
		var u Domain.BloodUnit
		if err := rows.Scan(
			&u.BloodUnitID, &u.DonationID, &u.BloodType, &u.ComponentType,
			&u.VolumeML, &u.CollectionDate, &u.ExpirationDate, &u.Status,
			&u.ReservedForHospitalID, &u.ReservedAt, &u.RequestID, &u.CreatedAt,
		); err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, nil
}

// DeleteWithAudit deletes a blood unit ONLY if its status is EXPIRED or USED.
// It writes an audit record first so analytics are preserved.
func (r *BloodInventoryRepository) DeleteWithAudit(unitID string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Fetch the unit — verify it exists and status is deletable
	var bloodType string
	var volumeML int
	var status string
	err = tx.QueryRow(`
		SELECT blood_type, volume_ml, status FROM blood_units WHERE blood_unit_id = $1
	`, unitID).Scan(&bloodType, &volumeML, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("blood unit not found")
		}
		return err
	}

	if status != "EXPIRED" && status != "USED" {
		return fmt.Errorf("cannot delete: unit status is '%s'. Only EXPIRED or USED units can be deleted", status)
	}

	// Write audit record before deletion
	_, err = tx.Exec(`
		INSERT INTO inventory_audit (blood_unit_id, blood_type, volume_ml, status_at_deletion)
		VALUES ($1, $2, $3, $4)
	`, unitID, bloodType, volumeML, status)
	if err != nil {
		return err
	}

	// Delete the unit
	_, err = tx.Exec(`DELETE FROM blood_units WHERE blood_unit_id = $1`, unitID)
	if err != nil {
		return err
	}

	return tx.Commit()
}