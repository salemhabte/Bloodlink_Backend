package Usecase

import (
	"bloodlink/Domain"
	"bloodlink/Repository"
	"time"

	"github.com/google/uuid"
)

type DonorBadgeUsecase struct {
	repo *Repository.DonorBadgeRepository
}

func NewDonorBadgeUsecase(repo *Repository.DonorBadgeRepository) *DonorBadgeUsecase {
	return &DonorBadgeUsecase{repo: repo}
}

// ================= BADGES =================

func (u *DonorBadgeUsecase) GetBadges(donorID string) ([]Domain.DonorBadge, error) {
	return u.repo.GetBadgesByDonor(donorID)
}

func (u *DonorBadgeUsecase) GetAllBadges() ([]Domain.DonorBadge, error) {
	return u.repo.GetAllBadges()
}

func (u *DonorBadgeUsecase) EvaluateBadges(donorID string) error {
	count, err := u.repo.CountDonationsByDonor(donorID)
	if err != nil {
		return err
	}

	rules := []struct {
		Count int
		Name  string
		Desc  string
	}{
		{1, "First Time Donor", "Completed first donation"},
	{5, "Regular Donor", "Donated at least 5 times"},
	{10, "Hero Donor", "Donated at least 10 times"},
	{20, "Champion Donor", "Donated at least 20 times"},
	{30, "Elite Donor", "Donated at least 30 times"},
	{50, "Legend Donor", "Donated at least 50 times"},
	{100, "Ultimate Hero", "Donated at least 100 times"},
	}

	for _, r := range rules {
		if count >= r.Count {

			exists, _ := u.repo.BadgeExists(donorID, r.Name)
			if exists {
				continue
			}

			badge := &Domain.DonorBadge{
				BadgeID:     uuid.New().String(),
				DonorID:     donorID,
				BadgeName:   r.Name,
				Description: r.Desc,
				AwardedAt:   time.Now(),
			}

			if err := u.repo.CreateBadge(badge); err != nil {
				return err
			}
		}
	}

	return nil
}

// ================= LEADERBOARD =================

func (u *DonorBadgeUsecase) GetLeaderboard(limit int) ([]Domain.LeaderboardEntry, error) {
	return u.repo.GetLeaderboard(limit)
}