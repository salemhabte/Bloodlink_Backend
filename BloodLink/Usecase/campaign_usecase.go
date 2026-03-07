package Usecase

import (
	"bloodlink/Domain"
	
)

type CampaignUsecase struct {
    Repo Domain.ICampaignRepository
}

func NewCampaignUsecase(repo Domain.ICampaignRepository) *CampaignUsecase {
    return &CampaignUsecase{Repo: repo}
}

func (u *CampaignUsecase) CreateCampaign(campaign *Domain.Campaign) error {
    return u.Repo.CreateCampaign(campaign)
}

func (u *CampaignUsecase) GetAllCampaigns() ([]Domain.Campaign, error) {
    return u.Repo.GetAllCampaigns()
}

func (u *CampaignUsecase) GetCampaignByID(id string) (*Domain.Campaign, error) {
    return u.Repo.GetCampaignByID(id)
}

func (u *CampaignUsecase) UpdateCampaign(campaign *Domain.Campaign) error {
    return u.Repo.UpdateCampaign(campaign)
}

func (u *CampaignUsecase) DeleteCampaign(id string) error {
    return u.Repo.DeleteCampaign(id)
}