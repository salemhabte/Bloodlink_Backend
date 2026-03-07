package Domain

type ICampaignRepository interface {
	CreateCampaign(campaign *Campaign) error
	GetAllCampaigns() ([]Campaign, error)
	GetCampaignByID(id string) (*Campaign, error)
	UpdateCampaign(campaign *Campaign) error
	DeleteCampaign(id string) error
}