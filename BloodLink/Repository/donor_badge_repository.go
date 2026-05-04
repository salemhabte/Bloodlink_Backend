package Repository

import (
	"bloodlink/Domain"
	"database/sql"
)

type DonorBadgeRepository struct {
	db *sql.DB
}

func NewDonorBadgeRepository(db *sql.DB) *DonorBadgeRepository {
	return &DonorBadgeRepository{db: db}
}

// ================= BADGES =================

func (r *DonorBadgeRepository) CreateBadge(b *Domain.DonorBadge) error {
	query := `
	INSERT INTO donor_badges (badge_id, donor_id, badge_name, description, awarded_at)
	VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query, b.BadgeID, b.DonorID, b.BadgeName, b.Description, b.AwardedAt)
	return err
}

func (r *DonorBadgeRepository) GetBadgesByDonor(donorID string) ([]Domain.DonorBadge, error) {
	query := `
	SELECT badge_id, donor_id, badge_name, description, awarded_at
	FROM donor_badges
	WHERE donor_id = $1
	ORDER BY awarded_at DESC
	`

	rows, err := r.db.Query(query, donorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var badges []Domain.DonorBadge

	for rows.Next() {
		var b Domain.DonorBadge
		if err := rows.Scan(&b.BadgeID, &b.DonorID, &b.BadgeName, &b.Description, &b.AwardedAt); err != nil {
			return nil, err
		}
		badges = append(badges, b)
	}

	return badges, nil
}

func (r *DonorBadgeRepository) GetAllBadges() ([]Domain.DonorBadge, error) {
	query := `
	SELECT badge_id, donor_id, badge_name, description, awarded_at
	FROM donor_badges
	ORDER BY awarded_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var badges []Domain.DonorBadge

	for rows.Next() {
		var b Domain.DonorBadge
		if err := rows.Scan(&b.BadgeID, &b.DonorID, &b.BadgeName, &b.Description, &b.AwardedAt); err != nil {
			return nil, err
		}
		badges = append(badges, b)
	}

	return badges, nil
}

// 🔥 ONLY CLEARED donations
func (r *DonorBadgeRepository) CountDonationsByDonor(donorID string) (int, error) {
	query := `
	SELECT COUNT(*)
	FROM donor_test_results
	WHERE donor_id = $1 AND overall_status = 'CLEARED'
	`

	var count int
	err := r.db.QueryRow(query, donorID).Scan(&count)
	return count, err
}

func (r *DonorBadgeRepository) BadgeExists(donorID, badgeName string) (bool, error) {
	query := `
	SELECT EXISTS (
		SELECT 1 FROM donor_badges
		WHERE donor_id = $1 AND badge_name = $2
	)
	`

	var exists bool
	err := r.db.QueryRow(query, donorID, badgeName).Scan(&exists)
	return exists, err
}

// ================= LEADERBOARD =================

func (r *DonorBadgeRepository) GetLeaderboard(limit int) ([]Domain.LeaderboardEntry, error) {
	query := `
	SELECT 
		donor_id,
		COUNT(*) AS donation_count
	FROM donor_test_results
	WHERE overall_status = 'CLEARED'
	GROUP BY donor_id
	ORDER BY donation_count DESC
	LIMIT $1
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Domain.LeaderboardEntry
	rank := 1

	for rows.Next() {
		var l Domain.LeaderboardEntry

		if err := rows.Scan(&l.DonorID, &l.DonationCount); err != nil {
			return nil, err
		}

		//  RANK HERE
		l.Rank = rank
		rank++

		result = append(result, l)
	}

	return result, nil
}