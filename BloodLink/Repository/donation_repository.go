package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
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
	sqlStr := `
	SELECT d.donor_id, d.user_id, u.full_name, u.email, u.phone, u.address, d.blood_type, d.status
	FROM donors d
	JOIN users u ON d.user_id = u.user_id
	WHERE LOWER(u.email) = LOWER(?) OR u.phone = ?
	LIMIT 1
	`

	var donor Domain.DonorResponse

	err := r.db.QueryRow(sqlStr, query, query).Scan(
		&donor.DonorID,
		&donor.UserID,
		&donor.FullName,
		&donor.Email,
		&donor.Phone,
		&donor.Address,
		&donor.BloodType,
		&donor.Status,
	)

	if err != nil {
		return nil, err
	}

	return &donor, nil
}


func (r *donationRepository) GetDonationByID(id string) (*Domain.DonationRecord, error) {

	query := `SELECT * FROM donation_records WHERE donation_id=?`

	row := r.db.QueryRow(query, id)

	var d Domain.DonationRecord

	err := row.Scan(
		&d.DonationID,
		&d.DonorID,
		&d.CollectedBy,
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

	query := `SELECT * FROM donation_records`

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
	SET weight=?, blood_pressure=?, hemoglobin=?, temperature=?, pulse=?, quantity_ml=?, collection_date=?
	WHERE donation_id=?`

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
	)

	return err
}
func (r *donationRepository) UpdateDonationStatus(donationID string, status string) error {
	query := `UPDATE donation_records SET status=? WHERE donation_id=?`
	_, err := r.db.Exec(query, status, donationID)
	return err
}