package Repository

import (
	"bloodlink/Domain"
	"database/sql"
)

type labAnalyticsRepository struct {
	db *sql.DB
}

func NewLabAnalyticsRepository(db *sql.DB) *labAnalyticsRepository {
	return &labAnalyticsRepository{db: db}
}
func (r *labAnalyticsRepository) GetLabDashboard(labID string) (*Domain.LabDashboard, error) {

	query := `
	SELECT
		-- total lab tests
		COUNT(*) AS total_tests,

		-- lab-specific results
		COALESCE(SUM(CASE WHEN overall_status = 'CLEARED' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN overall_status = 'TEMPORARILY_DEFERRED' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN overall_status = 'PERMANENTLY_DEFERRED' THEN 1 ELSE 0 END),0),

		COALESCE(SUM(CASE WHEN hiv_result = 'POSITIVE' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN hepatitis_result = 'POSITIVE' THEN 1 ELSE 0 END),0),
		COALESCE(SUM(CASE WHEN syphilis_result = 'POSITIVE' THEN 1 ELSE 0 END),0)

	FROM donor_test_results
	WHERE tested_by = $1
	`

	var d Domain.LabDashboard
	var temp, perm int

	err := r.db.QueryRow(query, labID).Scan(
		&d.TotalTests,
		&d.Cleared,
		&temp,
		&perm,
		&d.HIVPositive,
		&d.HepatitisPositive,
		&d.SyphilisPositive,
	)

	if err != nil {
		return nil, err
	}

	d.TemporarilyDeferred = temp
	d.PermanentlyDeferred = perm

	// percent calculation
	total := d.Cleared + d.TemporarilyDeferred + d.PermanentlyDeferred

	if total > 0 {
		d.ClearedPercent = float64(d.Cleared) / float64(total) * 100
		d.TemporarilyDeferredPercent = float64(d.TemporarilyDeferred) / float64(total) * 100
		d.PermanentlyDeferredPercent = float64(d.PermanentlyDeferred) / float64(total) * 100
	}

	return &d, nil
}
func (r *labAnalyticsRepository) GetPendingTests() (int, error) {

	query := `
	SELECT COUNT(*)
	FROM donation_records dr
	WHERE dr.status = 'APPROVED'
	AND NOT EXISTS (
		SELECT 1 FROM donor_test_results tr
		WHERE tr.donation_id = dr.donation_id
	)
	`

	var count int
	err := r.db.QueryRow(query).Scan(&count)

	return count, err
}