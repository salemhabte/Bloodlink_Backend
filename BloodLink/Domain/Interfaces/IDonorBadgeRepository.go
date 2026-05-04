package Domain

import (
	"bloodlink/Domain"
)

type IDonorBadgeRepository interface {
	CreateBadge(badge *Domain.DonorBadge) error
	GetBadgesByDonor(donorID string) ([]Domain.DonorBadge, error)
	GetAllBadges() ([]Domain.DonorBadge, error)

	CountDonationsByDonor(donorID string) (int, error)
	BadgeExists(donorID, badgeName string) (bool, error)

	GetLeaderboard(limit int) ([]Domain.LeaderboardEntry, error)
}