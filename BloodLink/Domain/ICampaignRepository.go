package Domain

type ICampaignRepository interface {
	CreateCampaign(campaign *Campaign) error
	GetAllCampaigns() ([]Campaign, error)// for both blood bank admin and the donor
	GetCampaignByID(id string) (*Campaign, error)
	UpdateCampaign(campaign *Campaign) error
	DeleteCampaign(id string) error

	// Donor feature
	GetCampaignsByLocation(location string) ([]Campaign, error)
}