package Repository

import (
	"bloodlink/Domain"
	"database/sql"
)

type adminAnalyticsRepository struct {
	db *sql.DB
}

func NewAdminAnalyticsRepository(db *sql.DB) *adminAnalyticsRepository {
	return &adminAnalyticsRepository{db: db}
}
func (r *adminAnalyticsRepository) GetDonorStats() (int, int, int, int, error) {

	// total registered donors
	var totalDonors int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM donors`).Scan(&totalDonors)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// donors who have donation records
	var withRecords int
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT donor_id)
		FROM donation_records
	`).Scan(&withRecords)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// approved donors
	var approved int
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT donor_id)
		FROM donation_records
		WHERE status = 'APPROVED'
	`).Scan(&approved)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// temporarily rejected donors
	var rejected int
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT donor_id)
		FROM donation_records
		WHERE status = 'TEMPORARILY_REJECTED'
	`).Scan(&rejected)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return totalDonors, withRecords, approved, rejected, nil
}
func (r *adminAnalyticsRepository) GetScreeningStats() (int, int, int, error) {

	query := `
	SELECT
		COALESCE(SUM(CASE WHEN overall_status = 'CLEARED' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN overall_status = 'TEMPORARILY_DEFERRED' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN overall_status = 'PERMANENTLY_DEFERRED' THEN 1 ELSE 0 END),0)
	FROM donor_test_results
	`

	var cleared, temp, perm int

	err := r.db.QueryRow(query).Scan(&cleared, &temp, &perm)
	if err != nil {
		return 0, 0, 0, err
	}

	return cleared, temp, perm, nil
}
func (r *adminAnalyticsRepository) GetCollectorStats() (int, int, []Domain.CollectorDonationStats, error) {

	// total collectors
	var totalCollectors int
	err := r.db.QueryRow(`
		SELECT COUNT(DISTINCT collected_by)
		FROM donation_records
	`).Scan(&totalCollectors)
	if err != nil {
		return 0, 0, nil, err
	}

	// total donations
	var totalDonations int
	err = r.db.QueryRow(`
		SELECT COUNT(*)
		FROM donation_records
	`).Scan(&totalDonations)
	if err != nil {
		return 0, 0, nil, err
	}

	// donations per collector
	query := `
	SELECT collected_by, COUNT(*)
	FROM donation_records
	WHERE collected_by IS NOT NULL
	GROUP BY collected_by
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return 0, 0, nil, err
	}
	defer rows.Close()

	var result []Domain.CollectorDonationStats

	for rows.Next() {
		var s Domain.CollectorDonationStats

		err := rows.Scan(&s.CollectorID, &s.TotalDonations)
		if err != nil {
			return 0, 0, nil, err
		}

		result = append(result, s)
	}

	return totalCollectors, totalDonations, result, nil
}
func (r *adminAnalyticsRepository) GetLabStats() (int, int, int, int, int, []Domain.LabTestStats, error) {

	// ================================
	// 1. TOTAL LAB TECHS
	// ================================
	var totalLabs int
	err := r.db.QueryRow(`
		SELECT COUNT(DISTINCT tested_by)
		FROM donor_test_results
		WHERE tested_by IS NOT NULL
	`).Scan(&totalLabs)
	if err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}

	// ================================
	// 2. TOTAL TEST RESULTS
	// ================================
	var totalTests int
	err = r.db.QueryRow(`
		SELECT COUNT(*)
		FROM donor_test_results
	`).Scan(&totalTests)
	if err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}

	// ================================
	// 3. STATUS COUNTS
	// ================================
	query := `
	SELECT
		COALESCE(SUM(CASE WHEN overall_status = 'CLEARED' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN overall_status = 'TEMPORARILY_DEFERRED' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN overall_status = 'PERMANENTLY_DEFERRED' THEN 1 ELSE 0 END),0)
	FROM donor_test_results
	`

	var cleared, temp, perm int

	err = r.db.QueryRow(query).Scan(&cleared, &temp, &perm)
	if err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}

	// ================================
	// 4. TESTS PER LAB TECH (NEW)
	// ================================
	queryPerLab := `
	SELECT tested_by, COUNT(*)
	FROM donor_test_results
	WHERE tested_by IS NOT NULL
	GROUP BY tested_by
	ORDER BY COUNT(*) DESC
	`

	rows, err := r.db.Query(queryPerLab)
	if err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}
	defer rows.Close()

	var perLab []Domain.LabTestStats

	for rows.Next() {
		var l Domain.LabTestStats

		err := rows.Scan(&l.LabTechID, &l.TotalTests)
		if err != nil {
			return 0, 0, 0, 0, 0, nil, err
		}

		perLab = append(perLab, l)
	}

	// check row iteration error
	if err = rows.Err(); err != nil {
		return 0, 0, 0, 0, 0, nil, err
	}

	// ================================
	// FINAL RETURN
	// ================================
	return totalLabs, totalTests, cleared, temp, perm, perLab, nil
}
func (r *adminAnalyticsRepository) GetInventoryStats() (int, []Domain.BloodTypeStat, int, error) {

	// ================================
	// 1. TOTAL BLOOD UNITS
	// ================================
	var total int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM blood_units
	`).Scan(&total)

	if err != nil {
		return 0, nil, 0, err
	}

	// ================================
	// 2. BLOOD TYPE DISTRIBUTION
	// ================================
	query := `
		SELECT blood_type, COUNT(*)
		FROM blood_units
		GROUP BY blood_type
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return 0, nil, 0, err
	}
	defer rows.Close()

	var bloodTypes []Domain.BloodTypeStat

	for rows.Next() {
		var b Domain.BloodTypeStat

		err := rows.Scan(&b.BloodType, &b.Count)
		if err != nil {
			return 0, nil, 0, err
		}

		// percentage
		if total > 0 {
			b.Percent = float64(b.Count) / float64(total) * 100
		}

		bloodTypes = append(bloodTypes, b)
	}

	// ================================
	// 3. NEAR EXPIRY (≤ 7 DAYS)
	// ================================
	var nearExpiry int
	err = r.db.QueryRow(`
		SELECT COUNT(*)
		FROM blood_units
		WHERE expiration_date <= CURRENT_DATE + INTERVAL '7 days'
		AND status = 'AVAILABLE'
	`).Scan(&nearExpiry)

	if err != nil {
		return 0, nil, 0, err
	}

	return total, bloodTypes, nearExpiry, nil
}