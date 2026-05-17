package Repository

import (
	"bloodlink/Domain"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"strings"
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
	err := r.DB.QueryRow("SELECT test_id FROM donor_test_results WHERE donation_id=$1", result.DonationID).Scan(&exists)
	if err == nil {
		return errors.New("test result for this donation already exists")
	}
	if err != sql.ErrNoRows {
		return err
	}

	// Insert new test result (no component_type — that's in blood_units)
	query := `
	INSERT INTO donor_test_results
	(test_id, donation_id, donor_id, tested_by, hiv_result, hepatitis_b_result, hepatitis_c_result, syphilis_result, blood_type, overall_status, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	bloodType := strings.ToUpper(strings.TrimSpace(result.BloodType))
	_, err = r.DB.Exec(query,
		result.TestID,
		result.DonationID,
		result.DonorID,
		result.TestedBy,
		result.HIVResult,
		result.HepatitisBResult,
		result.HepatitisCResult,
		result.SyphilisResult,
		bloodType,
		result.OverallStatus,
		time.Now(),
	)
	return err
}

func (r *LabRepository) CreateBloodUnit(unit *Domain.BloodUnit) error {
	query := `
	INSERT INTO blood_units
	(blood_unit_id, donation_id, blood_type, component_type, quantity_ml, collection_date, expiration_date, status, storage_location, rack_number, shelf_number)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.DB.Exec(query,
		unit.BloodUnitID,
		unit.DonationID,
		unit.BloodType,
		unit.ComponentType,
		unit.QuantityML,
		unit.CollectionDate,
		unit.ExpirationDate,
		unit.Status,
		unit.StorageLocation,
		unit.RackNumber,
		unit.ShelfNumber,
	)
	return err
}

func (r *LabRepository) UpdateDonorOverallStatus(donorID string, status string) error {
	query := `UPDATE donors SET overall_status=$1 WHERE donor_id=$2`
	_, err := r.DB.Exec(query, status, donorID)
	return err
}

func (r *LabRepository) UpdateDonorBloodType(donorID string, bloodType string) error {
	query := `UPDATE donors SET blood_type=$1 WHERE donor_id=$2`
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
		d.status,
		d.created_at
	FROM donation_records d
	JOIN donors dn ON d.donor_id = dn.donor_id
	JOIN users u ON dn.user_id = u.user_id
	JOIN users u2 ON d.collected_by = u2.user_id
	WHERE d.donation_id=$1
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
		&donation.Status,
		&donation.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("donation not found")
		}
		return nil, err
	}

	donation.OverallStatus = "PENDING"
	return &donation, nil
}

func (r *LabRepository) GetPendingDonationByID(donationID string) (*Domain.DonationRecord, error) {
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
		d.status,
		d.created_at
	FROM donation_records d
	JOIN donors dn ON d.donor_id = dn.donor_id
	JOIN users u ON dn.user_id = u.user_id
	JOIN users u2 ON d.collected_by = u2.user_id
	LEFT JOIN donor_test_results t ON d.donation_id = t.donation_id
	WHERE d.donation_id=$1 AND d.status = 'APPROVED' AND t.donation_id IS NULL
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
		&donation.Status,
		&donation.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("pending donation not found (already tested or rejected)")
		}
		return nil, err
	}

	donation.OverallStatus = "PENDING"
	return &donation, nil
}

func (r *LabRepository) GetTestResult(donationID string) (*Domain.DonorTestResult, error) {
	var result Domain.DonorTestResult
	
	query := `SELECT test_id, donation_id, donor_id, tested_by, hiv_result,
	          COALESCE(hepatitis_b_result, ''), COALESCE(hepatitis_c_result, ''),
	          syphilis_result, blood_type, overall_status, created_at
	          FROM donor_test_results WHERE donation_id=$1`
	
	err := r.DB.QueryRow(query, donationID).Scan(
		&result.TestID, &result.DonationID, &result.DonorID, &result.TestedBy,
		&result.HIVResult, &result.HepatitisBResult, &result.HepatitisCResult,
		&result.SyphilisResult,
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
			&d.DonorName,
			&d.CollectedBy,
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
		COALESCE(hepatitis_b_result, ''),
		COALESCE(hepatitis_c_result, ''),
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
			&rlt.HepatitisBResult,
			&rlt.HepatitisCResult,
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
		COALESCE(hepatitis_b_result, ''),
		COALESCE(hepatitis_c_result, ''),
		syphilis_result,
		blood_type,
		overall_status,
		created_at
	FROM donor_test_results
	WHERE overall_status = $1
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
			&rlt.HIVResult, &rlt.HepatitisBResult, &rlt.HepatitisCResult,
			&rlt.SyphilisResult,
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
	query := `
	UPDATE donor_test_results
	SET hiv_result=$1, hepatitis_b_result=$2, hepatitis_c_result=$3, syphilis_result=$4, overall_status=$5, blood_type=$6
	WHERE donation_id=$7
	`
	bloodType := strings.ToUpper(strings.TrimSpace(result.BloodType))
	res, err := r.DB.Exec(query,
		result.HIVResult,
		result.HepatitisBResult,
		result.HepatitisCResult,
		result.SyphilisResult,
		result.OverallStatus,
		bloodType,
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
	query := `UPDATE blood_units SET is_deleted = true WHERE donation_id = $1`
	_, err := r.DB.Exec(query, donationID)
	return err
}

// DeleteBloodUnitsByDonationID removes all blood units for a donation (used on update/re-creation)
func (r *LabRepository) DeleteBloodUnitsByDonationID(donationID string) error {
	query := `UPDATE blood_units SET is_deleted = true WHERE donation_id = $1`
	_, err := r.DB.Exec(query, donationID)
	return err
}

// GetBloodUnitByDonationID returns the first blood unit for a donation (backward compat)
func (r *LabRepository) GetBloodUnitByDonationID(donationID string) (*Domain.BloodUnit, error) {
	query := `SELECT blood_unit_id, donation_id, blood_type, COALESCE(component_type,''), quantity_ml, collection_date, expiration_date, status,
	          COALESCE(storage_location,''), COALESCE(rack_number,''), COALESCE(shelf_number,'')
	          FROM blood_units WHERE donation_id=$1 LIMIT 1`
	row := r.DB.QueryRow(query, donationID)

	var unit Domain.BloodUnit
	err := row.Scan(&unit.BloodUnitID, &unit.DonationID, &unit.BloodType, &unit.ComponentType,
		&unit.QuantityML, &unit.CollectionDate, &unit.ExpirationDate, &unit.Status,
		&unit.StorageLocation, &unit.RackNumber, &unit.ShelfNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &unit, nil
}

// GetBloodUnitsByDonationID returns ALL blood units for a donation
func (r *LabRepository) GetBloodUnitsByDonationID(donationID string) ([]Domain.BloodUnit, error) {
	query := `SELECT blood_unit_id, donation_id, blood_type, COALESCE(component_type,''), quantity_ml, collection_date, expiration_date, status,
	          COALESCE(storage_location,''), COALESCE(rack_number,''), COALESCE(shelf_number,'')
	          FROM blood_units WHERE donation_id=$1`
	rows, err := r.DB.Query(query, donationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []Domain.BloodUnit
	for rows.Next() {
		var u Domain.BloodUnit
		err := rows.Scan(&u.BloodUnitID, &u.DonationID, &u.BloodType, &u.ComponentType,
			&u.QuantityML, &u.CollectionDate, &u.ExpirationDate, &u.Status,
			&u.StorageLocation, &u.RackNumber, &u.ShelfNumber)
		if err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, nil
}

// Update blood unit
func (r *LabRepository) UpdateBloodUnit(unit *Domain.BloodUnit) error {
	query := `
	UPDATE blood_units 
	SET blood_type=$1, status=$2, storage_location=$3, rack_number=$4, shelf_number=$5
	WHERE donation_id=$6
	`
	_, err := r.DB.Exec(query, unit.BloodType, unit.Status, unit.StorageLocation, unit.RackNumber, unit.ShelfNumber, unit.DonationID)
	return err
}

func (r *LabRepository) GetTestResultsByLabTech(labTechID string) ([]Domain.DonorTestResult, error) {
	query := `
	SELECT 
		test_id,
		donation_id,
		donor_id,
		tested_by,
		hiv_result,
		COALESCE(hepatitis_b_result, ''),
		COALESCE(hepatitis_c_result, ''),
		syphilis_result,
		blood_type,
		overall_status,
		created_at
	FROM donor_test_results
	WHERE tested_by = $1
	`

	rows, err := r.DB.Query(query, labTechID)
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
			&rlt.HepatitisBResult,
			&rlt.HepatitisCResult,
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

func (r *LabRepository) FilterTestResults(filter Domain.TestFilter) ([]Domain.DonorTestResult, error) {

	query := `
	SELECT DISTINCT
		t.test_id,
		t.donation_id,
		t.donor_id,
		t.tested_by,
		t.hiv_result,
		COALESCE(t.hepatitis_b_result, ''),
		COALESCE(t.hepatitis_c_result, ''),
		t.syphilis_result,
		t.blood_type,
		t.overall_status,
		t.created_at
	FROM donor_test_results t
	LEFT JOIN blood_units bu ON t.donation_id = bu.donation_id
	WHERE 1=1
	`

	args := []interface{}{}
	i := 1

	if filter.LabTechID != "" {
		query += fmt.Sprintf(" AND t.tested_by = $%d", i)
		args = append(args, filter.LabTechID)
		i++
	}
	if filter.OverallStatus != "" {
		query += fmt.Sprintf(" AND UPPER(t.overall_status) = UPPER($%d)", i)
		args = append(args, filter.OverallStatus)
		i++
	}
	if filter.BloodType != "" {
		query += fmt.Sprintf(" AND UPPER(t.blood_type) = UPPER($%d)", i)
		args = append(args, filter.BloodType)
		i++
	}
	if filter.ComponentType != "" {
		query += fmt.Sprintf(" AND UPPER(bu.component_type) = UPPER($%d)", i)
		args = append(args, filter.ComponentType)
		i++
	}
	if filter.StorageLocation != "" {
		query += fmt.Sprintf(" AND UPPER(bu.storage_location) LIKE UPPER($%d)", i)
		args = append(args, "%"+filter.StorageLocation+"%")
		i++
	}
	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND t.created_at >= $%d", i)
		args = append(args, filter.StartDate)
		i++
	}
	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND t.created_at <= $%d", i)
		args = append(args, filter.EndDate)
		i++
	}

	query += " ORDER BY t.created_at DESC"

	rows, err := r.DB.Query(query, args...)
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
			&rlt.HepatitisBResult,
			&rlt.HepatitisCResult,
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

func (r *LabRepository) GetMyTestResultsFiltered(filter Domain.TestFilter) ([]Domain.DonorTestResult, error) {
	return r.FilterTestResults(filter)
}

// GetLatestTestResultByDonor returns the most recent test result for a donor,
// including the tester's full name and the campaign address (location) if applicable.
func (r *LabRepository) GetLatestTestResultByDonor(donorID string) (*Domain.DonorTestResult, error) {
	var result Domain.DonorTestResult

	query := `
	SELECT
		t.test_id,
		t.donation_id,
		t.donor_id,
		t.tested_by,
		u.full_name              AS tester_name,
		COALESCE(c.location, '') AS campaign_address,
		t.hiv_result,
		COALESCE(t.hepatitis_b_result, ''),
		COALESCE(t.hepatitis_c_result, ''),
		t.syphilis_result,
		t.blood_type,
		t.overall_status,
		t.created_at
	FROM donor_test_results t
	JOIN users u ON t.tested_by = u.user_id
	JOIN donation_records dr ON t.donation_id = dr.donation_id
	LEFT JOIN campaigns c ON dr.campaign_id = c.campaign_id
	WHERE t.donor_id = $1
	ORDER BY t.created_at DESC
	LIMIT 1
	`

	err := r.DB.QueryRow(query, donorID).Scan(
		&result.TestID, &result.DonationID, &result.DonorID, &result.TestedBy,
		&result.TesterName, &result.CampaignAddress,
		&result.HIVResult, &result.HepatitisBResult, &result.HepatitisCResult,
		&result.SyphilisResult,
		&result.BloodType, &result.OverallStatus, &result.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil if no test results are found
		}
		return nil, err
	}

	return &result, nil
}
