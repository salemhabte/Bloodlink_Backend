package Domain

import "bloodlink/Domain"

type ICampaignAnalyticsRepository interface {
	GetCampaignStats() (*Domain.CampaignStats, error)

	GetCampaignDonationStats(campaignID string) (*Domain.CampaignDonationStats, error)

	GetAllCampaignDonationStats() ([]Domain.CampaignDonationStats, error)
}