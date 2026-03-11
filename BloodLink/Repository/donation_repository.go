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
// SearchDonorByEmail finds a donor using the user's email
func (r *donationRepository) SearchDonorByEmail(email string) (*Domain.Donor, error) {

	query := `
	SELECT d.donor_id
	FROM donors d
	JOIN users u ON d.user_id = u.user_id
	WHERE u.email = ?
	`

	var donor Domain.Donor

	err := r.db.QueryRow(query, email).Scan(&donor.DonorID)

	if err != nil {
		return nil, err
	}

	return &donor, nil
}
// CreateDonation inserts a new donation record into the donation_records table
func (r *donationRepository) CreateDonation(record *Domain.DonationRecord) error {

	query := `
	INSERT INTO donation_records (
		donation_id,
		donor_id,
		collected_by,
		collection_date,
		weight,
		blood_pressure,
		hemoglobin,
		temperature,
		pulse,
		quantity_ml,
		status
	)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
	)

	return err
}
func (r *donationRepository) UpdateDonationStatus(donationID string, status string) error {
	query := `UPDATE donation_records SET status = ? WHERE donation_id = ?`
	_, err := r.db.Exec(query, status, donationID)
	return err
}