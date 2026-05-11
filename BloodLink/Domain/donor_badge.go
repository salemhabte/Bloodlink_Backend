package Domain

import "time"

type DonorBadge struct {
	BadgeID     string    `json:"badge_id"`
	DonorID     string    `json:"donor_id"`
	BadgeName   string    `json:"badge_name"`
	Description string    `json:"description"`
	AwardedAt   time.Time `json:"awarded_at"`
}

type LeaderboardEntry struct {
	Rank               int    `json:"rank"`
	DonorID            string `json:"donor_id"`
	FullName           string `json:"full_name"`
	ProfilePictureURL  string `json:"profile_picture_url"`
	DonationCount      int    `json:"donation_count"`
}