package Usecase

import (
	"bloodlink/Domain"
	"bloodlink/Repository"
	"fmt"
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
	}{
		{1, "Starter Donor"},
		{5, "Regular Donor"},
		{10, "Hero Donor"},
		{20, "Champion Donor"},
		{30, "Elite Donor"},
		{50, "Legend Donor"},
		{100, "Ultimate Hero"},
	}

	for _, r := range rules {

		if count >= r.Count {

			// Dynamic description
			description := fmt.Sprintf(
				"Completed %d successful donations",
				count,
			)

			exists, _ := u.repo.BadgeExists(donorID, r.Name)

			// If badge already exists -> update description
			if exists {

				err := u.repo.UpdateBadgeDescription(
					donorID,
					r.Name,
					description,
				)

				if err != nil {
					return err
				}

				continue
			}

			// Create new badge
			badge := &Domain.DonorBadge{
				BadgeID:     uuid.New().String(),
				DonorID:     donorID,
				BadgeName:   r.Name,
				Description: description,
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