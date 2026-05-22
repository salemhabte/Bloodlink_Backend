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
    donation_id, donor_id, campaign_id, collected_by, collection_date,
    weight, blood_pressure, hemoglobin, temperature, pulse,
    quantity_ml, status, rejection_reason
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
`

	var campaignID interface{}
if record.CampaignID != nil {
    campaignID = *record.CampaignID
} else {
    campaignID = nil
}

_, err := r.db.Exec(
    query,
    record.DonationID,      // $1
    record.DonorID,         // $2
    campaignID,             // $3 
    record.CollectedBy,     // $4
    record.CollectionDate,  // $5
    record.Weight,          // $6
    record.BloodPressure,   // $7
    record.Hemoglobin,      // $8
    record.Temperature,     // $9
    record.Pulse,           // $10
    record.QuantityML,      // $11
    record.Status,          // $12
    record.RejectionReason, // $13
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
		d.status,
		d.overall_status
	FROM donors d
	JOIN users u ON d.user_id = u.user_id
	WHERE LOWER(TRIM(u.email)) = LOWER($1)
	   OR u.phone LIKE '%' || $2 || '%'
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
		&donor.Status,
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
		d.campaign_id,
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
		COALESCE(d.rejection_reason, '') AS rejection_reason,
		COALESCE(c.title, '') AS campaign_title,
		COALESCE(c.location, '') AS campaign_address,
		d.created_at
	FROM donation_records d
	JOIN donors dn ON d.donor_id = dn.donor_id
	JOIN users u1 ON dn.user_id = u1.user_id
	JOIN users u2 ON d.collected_by = u2.user_id
	LEFT JOIN campaigns c ON d.campaign_id = c.campaign_id
	WHERE d.donation_id=$1
	`

	var d Domain.DonationRecord

	err := r.db.QueryRow(query, id).Scan(
		&d.DonationID,
		&d.DonorID,
		&d.CampaignID,
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
		&d.RejectionReason,
		&d.CampaignTitle,
		&d.CampaignAddress,
		&d.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (r *donationRepository) GetLastDonationByDonor(donorID string) (*Domain.DonationRecord, error) {

	query := `
	SELECT donation_id, donor_id, collection_date
	FROM donation_records
	WHERE donor_id=$1 AND status='APPROVED'
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
SET weight=$1, blood_pressure=$2, hemoglobin=$3, temperature=$4, pulse=$5, quantity_ml=$6, collection_date=$7, status=$8, rejection_reason=$9
WHERE donation_id=$10 AND donor_id=$11`

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
		record.RejectionReason,
		record.DonationID,
		record.DonorID,
	)

	return err
}
func (r *donationRepository) UpdateDonationStatus(donationID string, status string) error {
	query := `UPDATE donation_records SET status=$1 WHERE donation_id=$2`
	_, err := r.db.Exec(query, status, donationID)
	return err
}
func (r *donationRepository) UpdateDonorWeight(donorID string, weight float64) error {
	query := `UPDATE donors SET weight=$1 WHERE donor_id=$2`
	_, err := r.db.Exec(query, weight, donorID)
	return err
}

func (r *donationRepository) UpdateDonorOverallStatus(donorID string, status string) error {
	query := `UPDATE donors SET overall_status=$1 WHERE donor_id=$2`
	_, err := r.db.Exec(query, status, donorID)
	return err
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
	WHERE d.donor_id = $1
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
		LOWER(TRIM(u.email)) = LOWER($1)
		OR u.phone LIKE '%' || $2 || '%'
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

// GetAllDonationsByDonor returns all donation records for a donor,
// including collector name, campaign title and campaign address.
func (r *donationRepository) GetAllDonationsByDonor(donorID string) ([]Domain.DonationRecord, error) {

	query := `
	SELECT 
		d.donation_id,
		d.donor_id,
		d.campaign_id,
		d.collected_by,
		u1.full_name                AS donor_name,
		u2.full_name                AS collector_name,
		COALESCE(c.title, '')       AS campaign_title,
		COALESCE(c.location, '')    AS campaign_address,
		d.collection_date,
		d.weight,
		d.blood_pressure,
		d.hemoglobin,
		d.temperature,
		d.pulse,
		d.quantity_ml,
		d.status,
		COALESCE(d.rejection_reason, '') AS rejection_reason,
		d.created_at
	FROM donation_records d
	JOIN donors dn ON d.donor_id = dn.donor_id
	JOIN users u1 ON dn.user_id = u1.user_id
	JOIN users u2 ON d.collected_by = u2.user_id
	LEFT JOIN campaigns c ON d.campaign_id = c.campaign_id
	WHERE d.donor_id = $1
	ORDER BY d.collection_date DESC
	`

	rows, err := r.db.Query(query, donorID)
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
			&d.CampaignID,
			&d.CollectedBy,
			&d.DonorName,
			&d.CollectorName,
			&d.CampaignTitle,
			&d.CampaignAddress,
			&d.CollectionDate,
			&d.Weight,
			&d.BloodPressure,
			&d.Hemoglobin,
			&d.Temperature,
			&d.Pulse,
			&d.QuantityML,
			&d.Status,
			&d.RejectionReason,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		donations = append(donations, d)
	}

	return donations, nil
}
func (r *donationRepository) GetDonorOverallStatus(donorID string) (string, error) {

	query := `
	SELECT overall_status
	FROM donors
	WHERE donor_id = $1
	`

	var status string

	err := r.db.QueryRow(query, donorID).Scan(&status)
	if err != nil {
		return "", err
	}

	return status, nil
}

func (r *donationRepository) GetDonationsByCollector(collectorID string) ([]Domain.DonationRecord, error) {
	query := `
	SELECT 
		d.donation_id,
		d.donor_id,
		d.campaign_id,
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
		COALESCE(d.rejection_reason, '') AS rejection_reason,
		COALESCE(c.title, '') AS campaign_title,
		COALESCE(c.location, '') AS campaign_address,
		d.created_at
	FROM donation_records d
	JOIN donors dn ON d.donor_id = dn.donor_id
	JOIN users u1 ON dn.user_id = u1.user_id
	JOIN users u2 ON d.collected_by = u2.user_id
	LEFT JOIN campaigns c ON d.campaign_id = c.campaign_id
	WHERE d.collected_by = $1
	ORDER BY d.collection_date DESC
	`

	rows, err := r.db.Query(query, collectorID)
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
			&d.CampaignID,
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
			&d.RejectionReason,
			&d.CampaignTitle,
			&d.CampaignAddress,
			&d.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		donations = append(donations, d)
	}

	return donations, nil
}
func (r *donationRepository) GetDonations(filter Domain.DonationFilter) ([]Domain.DonationRecord, error) {

	query := `SELECT 
	d.donation_id,
	d.donor_id,
	d.campaign_id,
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
	COALESCE(d.rejection_reason, '') AS rejection_reason,

	COALESCE(tr.overall_status, dn.overall_status) AS overall_status,
	COALESCE(c.title, '') AS campaign_title,
	COALESCE(c.location, '') AS campaign_address,

	d.created_at

FROM donation_records d

LEFT JOIN donors dn ON d.donor_id = dn.donor_id
LEFT JOIN users u1 ON dn.user_id = u1.user_id
LEFT JOIN users u2 ON d.collected_by = u2.user_id
LEFT JOIN donor_test_results tr ON d.donation_id = tr.donation_id
LEFT JOIN campaigns c ON d.campaign_id = c.campaign_id

WHERE 1=1`

	var args []interface{}
	argIndex := 1

	// ================================
	// FILTER: collector
	// ================================
	if filter.CollectorID != "" {
	query += fmt.Sprintf(" AND d.collected_by = $%d", argIndex)
	args = append(args, filter.CollectorID)
	argIndex++
}

if filter.DonorID != "" {
	query += fmt.Sprintf(" AND d.donor_id = $%d", argIndex)
	args = append(args, filter.DonorID)
	argIndex++
}

if filter.Status != "" {
	query += fmt.Sprintf(" AND d.status = $%d", argIndex)
	args = append(args, filter.Status)
	argIndex++
}

if filter.StartDate != "" {
	query += fmt.Sprintf(" AND d.collection_date >= $%d", argIndex)
	args = append(args, filter.StartDate)
	argIndex++
}

if filter.EndDate != "" {
	query += fmt.Sprintf(" AND d.collection_date <= $%d", argIndex)
	args = append(args, filter.EndDate)
	argIndex++
}

	query += " ORDER BY d.collection_date DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Domain.DonationRecord

	for rows.Next() {
		var d Domain.DonationRecord

		err := rows.Scan(
			&d.DonationID,
			&d.DonorID,
			&d.CampaignID,
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
			&d.RejectionReason,
			&d.OverallStatus,
			&d.CampaignTitle,
			&d.CampaignAddress,
			&d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, d)
	}

	return result, nil
}