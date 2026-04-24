package Repository

import (
	"bloodlink/Domain"
	"database/sql"
)

type campaignAnalyticsRepository struct {
	db *sql.DB
}

func NewCampaignAnalyticsRepository(db *sql.DB) *campaignAnalyticsRepository {
	return &campaignAnalyticsRepository{db: db}
}

//
// 1. CAMPAIGN OVERVIEW STATS
//
func (r *campaignAnalyticsRepository) GetCampaignStats() (*Domain.CampaignStats, error) {

	query := `
	SELECT
		COUNT(*) AS total,
		COALESCE(SUM(CASE WHEN NOW() BETWEEN start_date AND end_date THEN 1 ELSE 0 END),0) AS active,
		COALESCE(SUM(CASE WHEN end_date < NOW() THEN 1 ELSE 0 END),0) AS completed,
		COALESCE(SUM(CASE WHEN start_date > NOW() THEN 1 ELSE 0 END),0) AS upcoming
	FROM campaigns;
	`

	var stats Domain.CampaignStats

	err := r.db.QueryRow(query).Scan(
		&stats.TotalCampaigns,
		&stats.ActiveCampaigns,
		&stats.CompletedCampaigns,
		&stats.UpcomingCampaigns,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

//
// 2. SINGLE CAMPAIGN ANALYTICS
//
func (r *campaignAnalyticsRepository) GetCampaignDonationStats(campaignID string) (*Domain.CampaignDonationStats, error) {

	query := `
	SELECT
		COUNT(*) AS total_donations,
		COALESCE(SUM(quantity_ml),0) AS total_ml,
		COALESCE(COUNT(DISTINCT donor_id),0) AS unique_donors,
		COALESCE(SUM(CASE WHEN status = 'APPROVED' THEN 1 ELSE 0 END),0) AS approved,

	COALESCE(SUM(CASE WHEN status = 'TEMPORARILY_REJECTED' THEN 1 ELSE 0 END),0) AS temporarily_rejected
	FROM donation_records
	WHERE campaign_id = $1;
	`

	var s Domain.CampaignDonationStats
	s.CampaignID = campaignID

	var uniqueDonors int

	err := r.db.QueryRow(query, campaignID).Scan(
		&s.TotalDonations,
		&s.TotalBloodML,
		&uniqueDonors,
		&s.ApprovedCount,
		&s.TemporarilyRejectedCount,
	)

	if err != nil {
		return nil, err
	}

	//  Derived calculations (safe)
	if s.TotalDonations > 0 {
		s.SuccessRate = float64(s.ApprovedCount) / float64(s.TotalDonations) * 100
		s.DropOffRate = float64(s.TemporarilyRejectedCount) / float64(s.TotalDonations) * 100
	}

	if uniqueDonors > 0 {
		s.AvgDonationPerDonor = float64(s.TotalDonations) / float64(uniqueDonors)
	}

	return &s, nil
}

//
// 3. ALL CAMPAIGNS ANALYTICS
//
func (r *campaignAnalyticsRepository) GetAllCampaignDonationStats() ([]Domain.CampaignDonationStats, error) {

	query := `
	SELECT
		campaign_id,
		COUNT(*) AS total_donations,
		COALESCE(SUM(quantity_ml),0) AS total_ml,
		COALESCE(COUNT(DISTINCT donor_id),0) AS unique_donors,
		COALESCE(SUM(CASE WHEN status = 'APPROVED' THEN 1 ELSE 0 END),0) AS approved,

	COALESCE(SUM(CASE WHEN status = 'TEMPORARILY_REJECTED' THEN 1 ELSE 0 END),0) AS temporarily_rejected
	FROM donation_records
	WHERE campaign_id IS NOT NULL
	GROUP BY campaign_id;
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Domain.CampaignDonationStats

	for rows.Next() {
		var s Domain.CampaignDonationStats
		var uniqueDonors int

		err := rows.Scan(
			&s.CampaignID,
			&s.TotalDonations,
			&s.TotalBloodML,
			&uniqueDonors,
			&s.ApprovedCount,
			&s.TemporarilyRejectedCount,
		)

		if err != nil {
			return nil, err
		}

		// Safe calculations
		if s.TotalDonations > 0 {
			s.SuccessRate = float64(s.ApprovedCount) / float64(s.TotalDonations) * 100
			s.DropOffRate = float64(s.TemporarilyRejectedCount) / float64(s.TotalDonations) * 100
		}

		if uniqueDonors > 0 {
			s.AvgDonationPerDonor = float64(s.TotalDonations) / float64(uniqueDonors)
		}

		result = append(result, s)
	}

	return result, nil
}