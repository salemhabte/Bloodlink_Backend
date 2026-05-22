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
func (r *BloodInventoryRepository) GetAllBloodUnits(filter Domain.BloodUnitFilter) ([]Domain.BloodUnit, error) {
	query := `
	SELECT blood_unit_id, '', blood_type, COALESCE(component_type,''),
	       quantity_ml, collection_date, expiration_date, status,
	       COALESCE(reserved_for_hospital_id,''), reserved_at, COALESCE(request_id,''), created_at, is_deleted,
		   COALESCE(storage_location, ''), COALESCE(rack_number, ''), COALESCE(shelf_number, ''), COALESCE(position_number, '')
	FROM blood_units
	WHERE is_deleted = false
	`
	args := []interface{}{}
	placeholderID := 1

	if filter.BloodType != "" {
		query += fmt.Sprintf(" AND UPPER(blood_type) = UPPER($%d)", placeholderID)
		args = append(args, filter.BloodType)
		placeholderID++
	}
	if filter.ComponentType != "" {
		query += fmt.Sprintf(" AND UPPER(component_type) = UPPER($%d)", placeholderID)
		args = append(args, filter.ComponentType)
		placeholderID++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND UPPER(status) = UPPER($%d)", placeholderID)
		args = append(args, filter.Status)
		placeholderID++
	}
	if filter.Quantity > 0 {
		query += fmt.Sprintf(" AND quantity_ml >= $%d", placeholderID)
		args = append(args, filter.Quantity)
		placeholderID++
	}
	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND collection_date >= $%d", placeholderID)
		args = append(args, filter.StartDate)
		placeholderID++
	}
	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND collection_date <= $%d", placeholderID)
		args = append(args, filter.EndDate)
		placeholderID++
	}
	if filter.NearExpired {
		query += " AND expiration_date > CURRENT_DATE AND expiration_date <= CURRENT_DATE + 7"
	}

	query += " ORDER BY expiration_date ASC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []Domain.BloodUnit
	for rows.Next() {
		var u Domain.BloodUnit
		err := rows.Scan(
			&u.BloodUnitID, &u.DonationID, &u.BloodType, &u.ComponentType,
			&u.QuantityML, &u.CollectionDate, &u.ExpirationDate, &u.Status,
			&u.ReservedForHospitalID, &u.ReservedAt, &u.RequestID, &u.CreatedAt, &u.IsDeleted,
			&u.StorageLocation, &u.RackNumber, &u.ShelfNumber, &u.PositionNumber,
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
	       quantity_ml, collection_date, expiration_date, status,
	       COALESCE(reserved_for_hospital_id,''), reserved_at, COALESCE(request_id,''), created_at, is_deleted,
		   COALESCE(storage_location, ''), COALESCE(rack_number, ''), COALESCE(shelf_number, ''), COALESCE(position_number, '')
	FROM blood_units WHERE blood_unit_id = $1 AND is_deleted = false
	`
	var u Domain.BloodUnit
	err := r.DB.QueryRow(query, id).Scan(
		&u.BloodUnitID, &u.DonationID, &u.BloodType, &u.ComponentType,
		&u.QuantityML, &u.CollectionDate, &u.ExpirationDate, &u.Status,
		&u.ReservedForHospitalID, &u.ReservedAt, &u.RequestID, &u.CreatedAt, &u.IsDeleted,
		&u.StorageLocation, &u.RackNumber, &u.ShelfNumber, &u.PositionNumber,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// UpdateBloodUnitStatus updates the status of a blood unit
func (r *BloodInventoryRepository) UpdateBloodUnitStatus(id string, status string) error {
	query := `UPDATE blood_units SET status=$1 WHERE blood_unit_id=$2 AND is_deleted = false`
	_, err := r.DB.Exec(query, status, id)
	return err
}

// DeleteBloodUnitByID deletes a blood unit (legacy, no audit)
func (r *BloodInventoryRepository) DeleteBloodUnitByID(id string) error {
	query := `UPDATE blood_units SET is_deleted = true WHERE blood_unit_id=$1`
	_, err := r.DB.Exec(query, id)
	return err
}

// GetFullBloodUnitDetails returns enriched details for one unit
func (r *BloodInventoryRepository) GetFullBloodUnitDetails(id string) (map[string]interface{}, error) {
	query := `
SELECT
    bu.blood_unit_id, bu.blood_type, bu.quantity_ml,
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
		&bloodUnit.BloodUnitID, &bloodUnit.BloodType, &bloodUnit.QuantityML,
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
	SELECT hiv_result, hepatitis_b_result, hepatitis_c_result, syphilis_result
	FROM donor_test_results WHERE donation_id = $1
	`, donation["donation_id"])
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []map[string]string
	for rows.Next() {
		var hiv, hepB, hepC, syph string
		rows.Scan(&hiv, &hepB, &hepC, &syph)
		tests = append(tests, map[string]string{
			"hiv": hiv, "hepatitis_b": hepB, "hepatitis_c": hepC, "syphilis": syph,
		})
	}

	result["blood_unit"] = bloodUnit
	result["donor"] = donor
	result["donation"] = donation
	result["test_results"] = tests
	return result, nil
}

// FilterBloodUnits filters blood units by various criteria
func (r *BloodInventoryRepository) FilterBloodUnits(filter Domain.BloodUnitFilter) ([]Domain.BloodUnit, error) {

	query := `
	SELECT blood_unit_id, donation_id, blood_type, COALESCE(component_type,''),
	       quantity_ml, collection_date, expiration_date, status,
	       COALESCE(reserved_for_hospital_id,''), reserved_at, COALESCE(request_id,''), created_at, is_deleted,
		   COALESCE(storage_location, ''), COALESCE(rack_number, ''), COALESCE(shelf_number, ''), COALESCE(position_number, '')
	FROM blood_units
	WHERE is_deleted = false
	`
	args := []interface{}{}
	placeholderID := 1

	if filter.BloodType != "" {
		query += fmt.Sprintf(" AND blood_type = $%d", placeholderID)
		args = append(args, filter.BloodType)
		placeholderID++
	}
	if filter.ComponentType != "" {
		query += fmt.Sprintf(" AND component_type = $%d", placeholderID)
		args = append(args, filter.ComponentType)
		placeholderID++
	}
	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", placeholderID)
		args = append(args, filter.Status)
		placeholderID++
	}
	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND collection_date >= $%d", placeholderID)
		args = append(args, filter.StartDate)
		placeholderID++
	}
	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND collection_date <= $%d", placeholderID)
		args = append(args, filter.EndDate)
		placeholderID++
	}

	query += " ORDER BY collection_date ASC"

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
			&u.QuantityML, &u.CollectionDate, &u.ExpirationDate, &u.Status,
			&u.ReservedForHospitalID, &u.ReservedAt, &u.RequestID, &u.CreatedAt, &u.IsDeleted,
			&u.StorageLocation, &u.RackNumber, &u.ShelfNumber, &u.PositionNumber,
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
	query := `SELECT COUNT(*) FROM blood_units WHERE blood_type = $1 AND status = 'AVAILABLE' AND expiration_date > NOW() AND is_deleted = false`
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
func (r *BloodInventoryRepository) ReserveUnitsForHospital(bloodType string, componentType string, quantity int, hospitalID string, requestID string) ([]Domain.BloodUnit, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()

	// Select AVAILABLE units ordered by nearest expiry (FIFO), lock rows
	selectQuery := `
	SELECT blood_unit_id, donation_id, blood_type, COALESCE(component_type,''),
	       quantity_ml, collection_date, expiration_date, status, created_at, is_deleted
	FROM blood_units
	WHERE UPPER(blood_type) = UPPER($1) 
	  AND (UPPER(component_type) = UPPER($2) OR UPPER(REPLACE(component_type, ' ', '_')) = UPPER(REPLACE($2, ' ', '_')) OR UPPER(REPLACE(component_type, '_', ' ')) = UPPER(REPLACE($2, '_', ' ')))
	  AND status = 'AVAILABLE' 
	  AND expiration_date > NOW()
	  AND is_deleted = false
	ORDER BY expiration_date ASC
	FOR UPDATE SKIP LOCKED
	`
	rows, err := tx.Query(selectQuery, bloodType, componentType)
	if err != nil {
		return nil, err
	}

	var units []Domain.BloodUnit
	accumulatedQuantity := 0
	
	for rows.Next() {
		var u Domain.BloodUnit
		if err := rows.Scan(
			&u.BloodUnitID, &u.DonationID, &u.BloodType, &u.ComponentType,
			&u.QuantityML, &u.CollectionDate, &u.ExpirationDate, &u.Status, &u.CreatedAt, &u.IsDeleted,
		); err != nil {
			rows.Close()
			return nil, err
		}
		
		units = append(units, u)
		accumulatedQuantity += u.QuantityML
		
		if accumulatedQuantity >= quantity {
			break // Stop once we have enough quantity
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
	       quantity_ml, collection_date, expiration_date, status,
	       COALESCE(reserved_for_hospital_id,''), reserved_at, COALESCE(request_id,''), created_at, is_deleted,
		   COALESCE(storage_location, ''), COALESCE(rack_number, ''), COALESCE(shelf_number, ''), COALESCE(position_number, '')
	FROM blood_units
	WHERE status = 'RESERVED' AND reserved_for_hospital_id = $1 AND is_deleted = false
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
			&u.QuantityML, &u.CollectionDate, &u.ExpirationDate, &u.Status,
			&u.ReservedForHospitalID, &u.ReservedAt, &u.RequestID, &u.CreatedAt, &u.IsDeleted,
			&u.StorageLocation, &u.RackNumber, &u.ShelfNumber, &u.PositionNumber,
		); err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, nil
}

func (r *BloodInventoryRepository) GetReservedUnitsByRequestID(requestID string) ([]Domain.BloodUnit, error) {
	query := `
	SELECT blood_unit_id, donation_id, blood_type, COALESCE(component_type,''),
	       quantity_ml, collection_date, expiration_date, status,
	       COALESCE(reserved_for_hospital_id,''), reserved_at, COALESCE(request_id,''), created_at, is_deleted,
		   COALESCE(storage_location, ''), COALESCE(rack_number, ''), COALESCE(shelf_number, ''), COALESCE(position_number, '')
	FROM blood_units
	WHERE status = 'RESERVED' AND request_id = $1 AND is_deleted = false
	ORDER BY expiration_date ASC
	`
	rows, err := r.DB.Query(query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []Domain.BloodUnit
	for rows.Next() {
		var u Domain.BloodUnit
		if err := rows.Scan(
			&u.BloodUnitID, &u.DonationID, &u.BloodType, &u.ComponentType,
			&u.QuantityML, &u.CollectionDate, &u.ExpirationDate, &u.Status,
			&u.ReservedForHospitalID, &u.ReservedAt, &u.RequestID, &u.CreatedAt, &u.IsDeleted,
			&u.StorageLocation, &u.RackNumber, &u.ShelfNumber, &u.PositionNumber,
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
		SELECT blood_type, quantity_ml, status FROM blood_units WHERE blood_unit_id = $1
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
		INSERT INTO inventory_audit (blood_unit_id, blood_type, quantity_ml, status_at_deletion)
		VALUES ($1, $2, $3, $4)
	`, unitID, bloodType, volumeML, status)
	if err != nil {
		return err
	}

	// Soft Delete the unit
	_, err = tx.Exec(`UPDATE blood_units SET is_deleted = true WHERE blood_unit_id = $1`, unitID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BloodInventoryRepository) ConvertPlasmaToCryo(plasmaUnitID string, cryo *Domain.BloodUnit, cryoPoor *Domain.BloodUnit) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Soft delete the original plasma unit
	res, err := tx.Exec(`UPDATE blood_units SET is_deleted = true WHERE blood_unit_id = $1 AND component_type = 'PLASMA' AND is_deleted = false AND status = 'AVAILABLE'`, plasmaUnitID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("plasma unit not found or not eligible for conversion")
	}

	insertQuery := `
	INSERT INTO blood_units (
		blood_unit_id, donation_id, blood_type, component_type,
		quantity_ml, collection_date, expiration_date, status, created_at, storage_location, rack_number, shelf_number, position_number
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	// Insert Cryoprecipitate
	_, err = tx.Exec(insertQuery,
		cryo.BloodUnitID, cryo.DonationID, cryo.BloodType, cryo.ComponentType,
		cryo.QuantityML, cryo.CollectionDate, cryo.ExpirationDate, cryo.Status, cryo.CreatedAt,
		cryo.StorageLocation, cryo.RackNumber, cryo.ShelfNumber, cryo.PositionNumber,
	)
	if err != nil {
		return err
	}

	// Insert Cryo-poor Plasma
	_, err = tx.Exec(insertQuery,
		cryoPoor.BloodUnitID, cryoPoor.DonationID, cryoPoor.BloodType, cryoPoor.ComponentType,
		cryoPoor.QuantityML, cryoPoor.CollectionDate, cryoPoor.ExpirationDate, cryoPoor.Status, cryoPoor.CreatedAt,
		cryoPoor.StorageLocation, cryoPoor.RackNumber, cryoPoor.ShelfNumber, cryoPoor.PositionNumber,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BloodInventoryRepository) IsSlotOccupied(location, rack, shelf, position string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM blood_units 
		WHERE storage_location = $1 AND rack_number = $2 AND shelf_number = $3 AND position_number = $4
		AND status != 'USED' AND is_deleted = false
	`
	var count int
	err := r.DB.QueryRow(query, location, rack, shelf, position).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *BloodInventoryRepository) GetOccupiedSlotCount(location, rack, shelf string) (int, error) {
	query := `
		SELECT COUNT(*) FROM blood_units 
		WHERE storage_location = $1 AND rack_number = $2 AND shelf_number = $3
		AND status != 'USED' AND is_deleted = false
	`
	var count int
	err := r.DB.QueryRow(query, location, rack, shelf).Scan(&count)
	return count, err
}