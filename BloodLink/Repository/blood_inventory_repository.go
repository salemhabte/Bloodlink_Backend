package Repository

import (
	"bloodlink/Domain"
	"database/sql"
)

type BloodInventoryRepository struct {
	DB *sql.DB
}

func NewBloodInventoryRepository(db *sql.DB) *BloodInventoryRepository {
	return &BloodInventoryRepository{DB: db}
}

// 🔹 Get All
func (r *BloodInventoryRepository) GetAllBloodUnits() ([]Domain.BloodUnit, error) {
	query := `
	SELECT blood_unit_id, donation_id, blood_type, volume_ml,
	       collection_date, expiration_date, status, created_at
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
			&u.BloodUnitID,
			&u.DonationID,
			&u.BloodType,
			&u.VolumeML,
			&u.CollectionDate,
			&u.ExpirationDate,
			&u.Status,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		units = append(units, u)
	}

	return units, nil
}

// 🔹 Get By ID
func (r *BloodInventoryRepository) GetBloodUnitByID(id string) (*Domain.BloodUnit, error) {
	query := `
	SELECT blood_unit_id, donation_id, blood_type, volume_ml,
	       collection_date, expiration_date, status, created_at
	FROM blood_units WHERE blood_unit_id = ?
	`

	var u Domain.BloodUnit
	err := r.DB.QueryRow(query, id).Scan(
		&u.BloodUnitID,
		&u.DonationID,
		&u.BloodType,
		&u.VolumeML,
		&u.CollectionDate,
		&u.ExpirationDate,
		&u.Status,
		&u.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

// 🔹 Update Status
func (r *BloodInventoryRepository) UpdateBloodUnitStatus(id string, status string) error {
	query := `UPDATE blood_units SET status=? WHERE blood_unit_id=?`
	_, err := r.DB.Exec(query, status, id)
	return err
}

// 🔹 Delete
func (r *BloodInventoryRepository) DeleteBloodUnitByID(id string) error {
	query := `DELETE FROM blood_units WHERE blood_unit_id=?`
	_, err := r.DB.Exec(query, id)
	return err
}
func (r *BloodInventoryRepository) GetFullBloodUnitDetails(id string) (map[string]interface{}, error) {

	// 🔹 Blood Unit + Donation + Donor
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

WHERE bu.blood_unit_id = ?
`

	var result = make(map[string]interface{})

	row := r.DB.QueryRow(query, id)

	var bloodUnit Domain.BloodUnit
	var donationID, donorID, collectedBy string
	var donorName, donorEmail, donorPhone string

	err := row.Scan(
		&bloodUnit.BloodUnitID,
		&bloodUnit.BloodType,
		&bloodUnit.VolumeML,
		&bloodUnit.CollectionDate,
		&bloodUnit.ExpirationDate,
		&bloodUnit.Status,

		&donationID,
		&donorID,
		&collectedBy,

		&donorName,
		&donorEmail,
		&donorPhone,
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

	// 🔹 Get Lab Results
	rows, err := r.DB.Query(`
	SELECT hiv_result, hepatitis_result, syphilis_result
	FROM donor_test_results
	WHERE donation_id = ?
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
			"hiv":       hiv,
			"hepatitis": hep,
			"syphilis":  syph,
		})
	}

	result["blood_unit"] = bloodUnit
	result["donor"] = donor
	result["donation"] = donation
	result["test_results"] = tests

	return result, nil
}
func (r *BloodInventoryRepository) FilterBloodUnits(
	unitID, bloodType, status string,
	startDate, endDate string,
) ([]Domain.BloodUnit, error) {

	query := `
	SELECT blood_unit_id, donation_id, blood_type, volume_ml,
	       collection_date, expiration_date, status, created_at
	FROM blood_units
	WHERE 1=1
	`

	args := []interface{}{}

	// Filter by ID
	if unitID != "" {
		query += " AND blood_unit_id LIKE ?"
		args = append(args, "%"+unitID+"%")
	}

	//  Filter by blood type
	if bloodType != "" {
		query += " AND blood_type = ?"
		args = append(args, bloodType)
	}

	// Filter by status
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	// Date range (collection date)
	if startDate != "" && endDate != "" {
		query += " AND collection_date BETWEEN ? AND ?"
		args = append(args, startDate, endDate)
	}

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []Domain.BloodUnit

	for rows.Next() {
		var u Domain.BloodUnit
		rows.Scan(
			&u.BloodUnitID,
			&u.DonationID,
			&u.BloodType,
			&u.VolumeML,
			&u.CollectionDate,
			&u.ExpirationDate,
			&u.Status,
			&u.CreatedAt,
		)
		units = append(units, u)
	}

	return units, nil
}