package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
)

type collectorAnalyticsRepository struct {
	db *sql.DB
}

func NewCollectorAnalyticsRepository(db *sql.DB) Interfaces.ICollectorAnalyticsRepository {
	return &collectorAnalyticsRepository{db: db}
}

func (r *collectorAnalyticsRepository) GetCollectorKPI(collectorID string) (*Domain.CollectorKPI, error) {

	query := `
	SELECT
		COUNT(*) AS total,

		COALESCE(SUM(CASE 
			WHEN DATE(collection_date) = CURRENT_DATE THEN 1 
			ELSE 0 
		END), 0),

		COALESCE(SUM(CASE 
			WHEN status='APPROVED' THEN 1 
			ELSE 0 
		END),0),

		COALESCE(SUM(CASE 
			WHEN status='REJECTED_TEMPORARY' THEN 1 
			ELSE 0 
		END),0)

	FROM donation_records
	WHERE collected_by = $1
	`

	var kpi Domain.CollectorKPI
	kpi.CollectorID = collectorID

	err := r.db.QueryRow(query, collectorID).Scan(
		&kpi.TotalDonations,
		&kpi.TodaysDonations, 
		&kpi.ApprovedCount,
		&kpi.TemporarilyRejectedCount,
	)

	if err != nil {
		return nil, err
	}

	if kpi.TotalDonations > 0 {
		kpi.ApprovalRate = float64(kpi.ApprovedCount) / float64(kpi.TotalDonations) * 100
		kpi.RejectionRate = float64(kpi.TemporarilyRejectedCount) / float64(kpi.TotalDonations) * 100
	}

	return &kpi, nil
}
func (r *collectorAnalyticsRepository) GetTodayStats(collectorID string) (*Domain.CollectorTodayStats, error) {

	query := `
	SELECT
		COUNT(*) AS todays_donations,

		COALESCE(SUM(CASE 
			WHEN status = 'APPROVED' THEN 1 
			ELSE 0 
		END), 0) AS approved_today,

		COALESCE(SUM(CASE 
			WHEN status = 'REJECTED_TEMPORARY' THEN 1 
			ELSE 0 
		END), 0) AS rejected_today,

		(
			SELECT COUNT(*)
			FROM donors d
			WHERE NOT EXISTS (
				SELECT 1 
				FROM donation_records dr 
				WHERE dr.donor_id = d.donor_id
			)
		) AS pending_donors

	FROM donation_records
	WHERE collected_by = $1
	AND DATE(collection_date) = CURRENT_DATE
	`

	var s Domain.CollectorTodayStats

	err := r.db.QueryRow(query, collectorID).Scan(
		&s.TodaysDonations,
		&s.ApprovedToday,
		&s.RejectedToday,
		&s.PendingDonors,
	)

	if err != nil {
		return nil, err
	}

	return &s, nil
}
func (r *collectorAnalyticsRepository) GetDonorInsights(collectorID string) (*Domain.CollectorDonorInsights, error) {

	var insights Domain.CollectorDonorInsights

	// New donors: donated exactly once with this collector
	newDonorsQuery := `
	SELECT COUNT(DISTINCT donor_id)
	FROM donation_records
	WHERE collected_by = $1
	  AND donor_id IN (
	    SELECT donor_id FROM donation_records
	    GROUP BY donor_id
	    HAVING COUNT(*) = 1
	  )
	`
	if err := r.db.QueryRow(newDonorsQuery, collectorID).Scan(&insights.NewDonors); err != nil {
		return nil, err
	}

	// Returning donors: donated more than once and at least once with this collector
	returningDonorsQuery := `
	SELECT COUNT(DISTINCT donor_id)
	FROM donation_records
	WHERE collected_by = $1
	  AND donor_id IN (
	    SELECT donor_id FROM donation_records
	    GROUP BY donor_id
	    HAVING COUNT(*) > 1
	  )
	`
	if err := r.db.QueryRow(returningDonorsQuery, collectorID).Scan(&insights.ReturningDonors); err != nil {
		return nil, err
	}

	// High-risk donors: have at least one REJECTED_TEMPORARY donation with this collector
	highRiskQuery := `
	SELECT COUNT(DISTINCT donor_id)
	FROM donation_records
	WHERE collected_by = $1
	  AND status = 'REJECTED_TEMPORARY'
	`
	if err := r.db.QueryRow(highRiskQuery, collectorID).Scan(&insights.HighRiskDonors); err != nil {
		return nil, err
	}

	// Most active donor for this collector
	mostActiveQuery := `
	SELECT donor_id
	FROM donation_records
	WHERE collected_by = $1
	GROUP BY donor_id
	ORDER BY COUNT(*) DESC
	LIMIT 1
	`
	// Ignore no-rows error; leave MostActiveDonor as empty string if none found
	_ = r.db.QueryRow(mostActiveQuery, collectorID).Scan(&insights.MostActiveDonor)

	return &insights, nil
}