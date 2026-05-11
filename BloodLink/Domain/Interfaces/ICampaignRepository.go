package Domain

import (
	domain "bloodlink/Domain"
)

type ICampaignRepository interface {
	CreateCampaign(campaign *domain.Campaign) error
	GetAllCampaigns(filter domain.CampaignFilter) ([]domain.Campaign, error) // for both blood bank admin and the donor
	GetCampaignByID(id string) (*domain.Campaign, error)
	UpdateCampaign(campaign *domain.Campaign) error
	DeleteCampaign(id string) error

	// Donor feature
	GetCampaignsByLocation(filter domain.CampaignFilter) ([]domain.Campaign, error)
}
