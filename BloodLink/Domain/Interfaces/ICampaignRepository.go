package Domain

import (
	domain "bloodlink/Domain"
)

type ICampaignRepository interface {
	CreateCampaign(campaign *domain.Campaign) error
	GetAllCampaigns(filter domain.CampaignFilter, liveOnly bool) ([]domain.Campaign, error)
	GetCampaignByID(id string) (*domain.Campaign, error)
	GetLiveCampaignByID(id string) (*domain.Campaign, error)
	UpdateCampaign(campaign *domain.Campaign) error
	DeleteCampaign(id string) error
}
