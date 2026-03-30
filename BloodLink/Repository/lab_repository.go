package Repository

import (
	"bloodlink/Domain"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type LabRepository struct {
	DB *sql.DB
}

func NewLabRepository(db *sql.DB) *LabRepository {
	return &LabRepository{DB: db}
}

func (r *LabRepository) CreateTestResult(result *Domain.DonorTestResult) error {
	// Check if test already exists
	var exists string
	err := r.DB.QueryRow("SELECT test_id FROM donor_test_results WHERE donation_id=?", result.DonationID).Scan(&exists)
	if err == nil {
		return errors.New("test result for this donation already exists")
	}
	if err != sql.ErrNoRows {
		return err
	}

	// Insert new test result
	query := `
	INSERT INTO donor_test_results
	(test_id, donation_id, donor_id, tested_by, hiv_result, hepatitis_result, syphilis_result, blood_type, overall_status, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err = r.DB.Exec(query,
		result.TestID,
		result.DonationID,
		result.DonorID,
		result.TestedBy,
		result.HIVResult,
		result.HepatitisResult,
		result.SyphilisResult,
		result.BloodType,
		result.OverallStatus,
		time.Now(),
	)
	return err
}

func (r *LabRepository) CreateBloodUnit(unit *Domain.BloodUnit) error {
	query := `
	INSERT INTO blood_units
	(blood_unit_id, donation_id, blood_type, volume_ml, collection_date, expiration_date, status)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.DB.Exec(query,
		unit.BloodUnitID,
		unit.DonationID,
		unit.BloodType,
		unit.VolumeML,
		unit.CollectionDate,
		unit.ExpirationDate,
		unit.Status,
	)
	return err
}

func (r *LabRepository) UpdateDonorOverallStatus(donorID string, status string) error {
	query := `UPDATE donors SET overall_status=? WHERE donor_id=?`
	_, err := r.DB.Exec(query, status, donorID)
	return err
}

func (r *LabRepository) UpdateDonorBloodType(donorID string, bloodType string) error {
	query := `UPDATE donors SET blood_type=? WHERE donor_id=?`
	_, err := r.DB.Exec(query, bloodType, donorID)
	return err
}

func (r *LabRepository) GetDonationByID(donationID string) (*Domain.DonationRecord, error) {
	var donation Domain.DonationRecord

	query := `
	SELECT 
		d.donation_id,
		d.donor_id,
		u.full_name,
		d.collected_by,
		u2.full_name,
		d.collection_date,
		d.weight,
		d.blood_pressure,
		d.hemoglobin,
		d.temperature,
		d.pulse,
		d.quantity_ml,
		d.status,      -- include DB status
		d.created_at
	FROM donation_records d
	JOIN donors dn ON d.donor_id = dn.donor_id
	JOIN users u ON dn.user_id = u.user_id
	JOIN users u2 ON d.collected_by = u2.user_id
	WHERE d.donation_id=? AND d.status='APPROVED' AND d.collected_by IS NOT NULL
	`

	err := r.DB.QueryRow(query, donationID).Scan(
		&donation.DonationID,
		&donation.DonorID,
		&donation.DonorName,
		&donation.CollectedBy,
		&donation.CollectorName,
		&donation.CollectionDate,
		&donation.Weight,
		&donation.BloodPressure,
		&donation.Hemoglobin,
		&donation.Temperature,
		&donation.Pulse,
		&donation.QuantityML,
		&donation.Status,      // map the DB status here
		&donation.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("donation not found or not ready for lab")
		}
		return nil, err
	}

	// Lab-specific field
	donation.OverallStatus = "PENDING"  // this is what the lab sees

	return &donation, nil
}




func (r *LabRepository) GetTestResult(donationID string) (*Domain.DonorTestResult, error) {
	var result Domain.DonorTestResult
	
	query := `SELECT test_id, donation_id, donor_id, tested_by, hiv_result, hepatitis_result, syphilis_result, blood_type, overall_status, created_at FROM donor_test_results WHERE donation_id=?`
	
	err := r.DB.QueryRow(query, donationID).Scan(
		&result.TestID, &result.DonationID, &result.DonorID, &result.TestedBy,
		&result.HIVResult, &result.HepatitisResult, &result.SyphilisResult,
		&result.BloodType, &result.OverallStatus, &result.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	return &result, nil
}

func (r *LabRepository) GetPendingDonations() ([]Domain.DonationRecord, error) {
	query := `
	SELECT 
		d.donation_id,
		d.donor_id,
		u.full_name,
		d.collected_by,
		u2.full_name,
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
	JOIN users u ON dn.user_id = u.user_id
	JOIN users u2 ON d.collected_by = u2.user_id
	LEFT JOIN donor_test_results t ON d.donation_id = t.donation_id
	WHERE d.status = 'APPROVED' AND t.donation_id IS NULL
	`

	rows, err := r.DB.Query(query)
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
			&d.DonorName,     // u.full_name
			&d.CollectedBy,
			&d.CollectorName, // u2.full_name
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
		// Set pending by default for lab testing
		d.OverallStatus = "PENDING"

		donations = append(donations, d)
	}

	return donations, nil
}
func (r *LabRepository) GetAllTestResults() ([]Domain.DonorTestResult, error) {
	query := `
	SELECT 
		test_id,
		donation_id,
		donor_id,
		tested_by,
		hiv_result,
		hepatitis_result,
		syphilis_result,
		blood_type,
		overall_status,
		created_at
	FROM donor_test_results
	`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Domain.DonorTestResult

	for rows.Next() {
		var rlt Domain.DonorTestResult

		err := rows.Scan(
			&rlt.TestID,
			&rlt.DonationID,
			&rlt.DonorID,
			&rlt.TestedBy,
			&rlt.HIVResult,
			&rlt.HepatitisResult,
			&rlt.SyphilisResult,
			&rlt.BloodType,
			&rlt.OverallStatus,
			&rlt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		results = append(results, rlt)
	}

	return results, nil
}
func (r *LabRepository) GetTestResultsByStatus(status string) ([]Domain.DonorTestResult, error) {
	query := `
	SELECT 
		test_id,
		donation_id,
		donor_id,
		tested_by,
		hiv_result,
		hepatitis_result,
		syphilis_result,
		blood_type,
		overall_status,
		created_at
	FROM donor_test_results
	WHERE overall_status = ?
	`

	rows, err := r.DB.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []Domain.DonorTestResult

	for rows.Next() {
		var rlt Domain.DonorTestResult
	
		err := rows.Scan(
			&rlt.TestID, &rlt.DonationID, &rlt.DonorID, &rlt.TestedBy,
			&rlt.HIVResult, &rlt.HepatitisResult, &rlt.SyphilisResult,
			&rlt.BloodType, &rlt.OverallStatus, &rlt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		
		
		results = append(results, rlt)
	}

	return results, nil
}

func (r *LabRepository) UpdateTestResult(result *Domain.DonorTestResult) error {
	// Update test result only if donation_id exists
	query := `
UPDATE donor_test_results
SET hiv_result=?, hepatitis_result=?, syphilis_result=?, overall_status=?, blood_type=?
WHERE donation_id=?
`

res, err := r.DB.Exec(query,
	result.HIVResult,
	result.HepatitisResult,
	result.SyphilisResult,
	result.OverallStatus,
	result.BloodType,
	result.DonationID,
)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	fmt.Println("Rows affected:", rowsAffected)

	if rowsAffected == 0 {
	fmt.Println("No change detected, but update is valid")
	return nil
}

	return nil
}

func (r *LabRepository) DeleteBloodUnit(donationID string) error {
	query := `DELETE FROM blood_units WHERE donation_id = ?`
	_, err := r.DB.Exec(query, donationID)
	return err
}
// Get blood unit by donation ID
func (r *LabRepository) GetBloodUnitByDonationID(donationID string) (*Domain.BloodUnit, error) {
	query := "SELECT blood_unit_id, donation_id, blood_type, volume_ml, collection_date, expiration_date, status FROM blood_units WHERE donation_id=?"
	row := r.DB.QueryRow(query, donationID)

	var unit Domain.BloodUnit
	err := row.Scan(&unit.BloodUnitID, &unit.DonationID, &unit.BloodType, &unit.VolumeML, &unit.CollectionDate, &unit.ExpirationDate, &unit.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &unit, nil
}

// Update blood unit
func (r *LabRepository) UpdateBloodUnit(unit *Domain.BloodUnit) error {
	query := `
	UPDATE blood_units 
	SET blood_type=?, status=?
	WHERE donation_id=?
	`
	_, err := r.DB.Exec(query, unit.BloodType, unit.Status, unit.DonationID)
	return err
}
