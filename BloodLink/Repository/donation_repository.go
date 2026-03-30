package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
	"fmt"
	"strings"
)

// donationRepository implements the IDonationRepository interface
type donationRepository struct {
	db *sql.DB
}

// NewDonationRepository creates a new repository instance
func NewDonationRepository(db *sql.DB) Interfaces.IDonationRepository {
	return &donationRepository{db: db}
}
func (r *donationRepository) CreateDonation(record *Domain.DonationRecord) error {
	query := `
	INSERT INTO donation_records (
		donation_id, donor_id, collected_by, collection_date,
		weight, blood_pressure, hemoglobin, temperature, pulse,
		quantity_ml, status, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(
		query,
		record.DonationID,
		record.DonorID,
		record.CollectedBy,
		record.CollectionDate,
		record.Weight,
		record.BloodPressure,
		record.Hemoglobin,
		record.Temperature,
		record.Pulse,
		record.QuantityML,
		record.Status,
		record.CreatedAt,
	)

	return err
}
// SearchDonor finds a donor using email or phone
func (r *donationRepository) SearchDonor(query string) (*Domain.DonorResponse, error) {
	query = strings.TrimSpace(query) // trim hidden spaces

	sqlStr := `
	SELECT 
		d.donor_id,
		d.user_id,
		u.full_name,
		u.email,
		u.phone,
		d.blood_type,
		d.overall_status
	FROM donors d
	JOIN users u ON d.user_id = u.user_id
	WHERE LOWER(TRIM(u.email)) = LOWER(?)
	   OR u.phone LIKE CONCAT('%', ?, '%')
	LIMIT 1
	`

	var donor Domain.DonorResponse

	err := r.db.QueryRow(sqlStr, strings.ToLower(query), query).Scan(
		&donor.DonorID,
		&donor.UserID,
		&donor.FullName,
		&donor.Email,
		&donor.Phone,
		&donor.BloodType,
		&donor.OverallStatus,	
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("donor not found")
		}
		return nil, err
	}

	return &donor, nil
}

func (r *donationRepository) GetDonationByID(id string) (*Domain.DonationRecord, error) {

	query := `
	SELECT 
		d.donation_id,
		d.donor_id,
		d.collected_by,
		u1.full_name AS donor_name,
		u2.full_name AS collector_name,
		d.collection_date,
		d.weight,
		d.blood_pressure,
		d.hemoglobin,
		d.temperature,
		d.pulse,
		d.quantity_ml,
		d.status,
		d.created_at
	FROM donation_records d
	JOIN donors dn ON d.donor_id = dn.donor_id
	JOIN users u1 ON dn.user_id = u1.user_id
	JOIN users u2 ON d.collected_by = u2.user_id
	WHERE d.donation_id=?
	`

	var d Domain.DonationRecord

	err := r.db.QueryRow(query, id).Scan(
		&d.DonationID,
		&d.DonorID,
		&d.CollectedBy,
		&d.DonorName,
		&d.CollectorName,
		&d.CollectionDate,
		&d.Weight,
		&d.BloodPressure,
		&d.Hemoglobin,
		&d.Temperature,
		&d.Pulse,
		&d.QuantityML,
		&d.Status,
		&d.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &d, nil
}
func (r *donationRepository) GetAllDonations() ([]Domain.DonationRecord, error) {

	query := `
	SELECT 
		d.donation_id,
		d.donor_id,
		d.collected_by,
		u1.full_name AS donor_name,
		u2.full_name AS collector_name,
		d.collection_date,
		d.weight,
		d.blood_pressure,
		d.hemoglobin,
		d.temperature,
		d.pulse,
		d.quantity_ml,
		d.status,
		d.created_at
	FROM donation_records d
	JOIN donors dn ON d.donor_id = dn.donor_id
	JOIN users u1 ON dn.user_id = u1.user_id
	JOIN users u2 ON d.collected_by = u2.user_id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var donations []Domain.DonationRecord

	for rows.Next() {
		var d Domain.DonationRecord

		err := rows.Scan(
			&d.DonationID,
			&d.DonorID,
			&d.CollectedBy,
			&d.DonorName,
			&d.CollectorName,
			&d.CollectionDate,
			&d.Weight,
			&d.BloodPressure,
			&d.Hemoglobin,
			&d.Temperature,
			&d.Pulse,
			&d.QuantityML,
			&d.Status,
			&d.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		donations = append(donations, d)
	}

	return donations, nil
}
func (r *donationRepository) GetLastDonationByDonor(donorID string) (*Domain.DonationRecord, error) {

	query := `
	SELECT donation_id, donor_id, collection_date
	FROM donation_records
	WHERE donor_id=?
	ORDER BY collection_date DESC
	LIMIT 1`

	row := r.db.QueryRow(query, donorID)

	var d Domain.DonationRecord

	err := row.Scan(&d.DonationID, &d.DonorID, &d.CollectionDate)

	if err != nil {
		return nil, err
	}

	return &d, nil
}
func (r *donationRepository) UpdateDonation(record *Domain.DonationRecord) error {

	query := `
UPDATE donation_records
SET weight=?, blood_pressure=?, hemoglobin=?, temperature=?, pulse=?, quantity_ml=?, collection_date=?, status=?
WHERE donation_id=? AND donor_id=?`

	_, err := r.db.Exec(
		query,
		record.Weight,
		record.BloodPressure,
		record.Hemoglobin,
		record.Temperature,
		record.Pulse,
		record.QuantityML,
		record.CollectionDate,
		record.Status,
		record.DonationID,
		record.DonorID,
	)

	return err
}
func (r *donationRepository) UpdateDonationStatus(donationID string, status string) error {
	query := `UPDATE donation_records SET status=? WHERE donation_id=?`
	_, err := r.db.Exec(query, status, donationID)
	return err
}
func (r *donationRepository) UpdateDonorWeight(donorID string, weight float64) error {
	query := `UPDATE donors SET weight=? WHERE donor_id=?`

	result, err := r.db.Exec(query, weight, donorID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no donor found with id %s", donorID)
	}

	return nil
}
func (r *donationRepository) GetPendingDonors() ([]Domain.DonorResponse, error) {

	query := `
	SELECT 
		d.donor_id,
		d.user_id,
		u.full_name,
		u.email,
		u.phone,
		d.blood_type,
		d.overall_status
	FROM donors d
	JOIN users u ON d.user_id = u.user_id
	WHERE d.donor_id NOT IN (
		SELECT donor_id FROM donation_records
	)
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var donors []Domain.DonorResponse

	for rows.Next() {
		var d Domain.DonorResponse

		err := rows.Scan(
			&d.DonorID,
			&d.UserID,
			&d.FullName,
			&d.Email,
			&d.Phone,
			&d.BloodType,
			&d.OverallStatus,
		)

		if err != nil {
			return nil, err
		}

		donors = append(donors, d)
	}

	return donors, nil
}
func (r *donationRepository) GetPendingDonorByID(donorID string) (*Domain.DonorResponse, error) {

	query := `
	SELECT 
		d.donor_id,
		d.user_id,
		u.full_name,
		u.email,
		u.phone,
		d.blood_type,
		d.overall_status
	FROM donors d
	JOIN users u ON d.user_id = u.user_id
	WHERE d.donor_id = ?
	AND d.donor_id NOT IN (
		SELECT donor_id FROM donation_records
	)
	`

	var d Domain.DonorResponse

	err := r.db.QueryRow(query, donorID).Scan(
		&d.DonorID,
		&d.UserID,
		&d.FullName,
		&d.Email,
		&d.Phone,
		&d.BloodType,
		&d.OverallStatus,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("donor not found or already screened")
		}
		return nil, err
	}

	return &d, nil
}
func (r *donationRepository) SearchPendingDonor(query string) (*Domain.DonorResponse, error) {

	query = strings.TrimSpace(query)

	sqlStr := `
	SELECT 
		d.donor_id,
		d.user_id,
		u.full_name,
		u.email,
		u.phone,
		d.blood_type,
		d.overall_status
	FROM donors d
	JOIN users u ON d.user_id = u.user_id
	WHERE (
		LOWER(TRIM(u.email)) = LOWER(?)
		OR u.phone LIKE CONCAT('%', ?, '%')
	)
	AND NOT EXISTS (
		SELECT 1 FROM donation_records dr 
		WHERE dr.donor_id = d.donor_id
	)
	LIMIT 1
	`

	var donor Domain.DonorResponse

	err := r.db.QueryRow(sqlStr, strings.ToLower(query), query).Scan(
		&donor.DonorID,
		&donor.UserID,
		&donor.FullName,
		&donor.Email,
		&donor.Phone,
		&donor.BloodType,
		&donor.OverallStatus,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pending donor not found")
		}
		return nil, err
	}

	return &donor, nil
}