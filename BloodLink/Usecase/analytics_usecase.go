package Usecase

import (
	"bloodlink/Domain"
	Interface "bloodlink/Domain/Interfaces"
)
//campain analytics

type CampaignAnalyticsUsecase struct {
	repo Interface.ICampaignAnalyticsRepository
}

func NewCampaignAnalyticsUsecase(r Interface.ICampaignAnalyticsRepository) *CampaignAnalyticsUsecase {
	return &CampaignAnalyticsUsecase{repo: r}
}

func (u *CampaignAnalyticsUsecase) GetDashboardStats() (*Domain.CampaignStats, error) {
	return u.repo.GetCampaignStats()
}

func (u *CampaignAnalyticsUsecase) GetCampaignReport(campaignID string) (*Domain.CampaignDonationStats, error) {
	return u.repo.GetCampaignDonationStats(campaignID)
}

func (u *CampaignAnalyticsUsecase) GetAllCampaignReports() ([]Domain.CampaignDonationStats, error) {
	return u.repo.GetAllCampaignDonationStats()
}

// bloodcollecter analysis

type CollectorAnalyticsUsecase struct {
	repo Interface.ICollectorAnalyticsRepository
}

func NewCollectorAnalyticsUsecase(r Interface.ICollectorAnalyticsRepository) *CollectorAnalyticsUsecase {
	return &CollectorAnalyticsUsecase{repo: r}
}

func (u *CollectorAnalyticsUsecase) GetCollectorKPI(collectorID string) (*Domain.CollectorKPI, error) {
	return u.repo.GetCollectorKPI(collectorID)
}

func (u *CollectorAnalyticsUsecase) GetTodayStats(collectorID string) (*Domain.CollectorTodayStats, error) {
	return u.repo.GetTodayStats(collectorID)
}

func (u *CollectorAnalyticsUsecase) GetDonorInsights(collectorID string) (*Domain.CollectorDonorInsights, error) {
	return u.repo.GetDonorInsights(collectorID)
}

//labtech analytics
type LabAnalyticsUsecase struct {
	repo Interface.ILabAnalyticsRepository
}

func NewLabAnalyticsUsecase(r Interface.ILabAnalyticsRepository) *LabAnalyticsUsecase {
	return &LabAnalyticsUsecase{repo: r}
}
func (u *LabAnalyticsUsecase) GetDashboard(labID string) (*Domain.LabDashboard, error) {

	data, err := u.repo.GetLabDashboard(labID)
	if err != nil {
		return nil, err
	}

	// inject GLOBAL pending tests
	pending, err := u.repo.(interface {
		GetPendingTests() (int, error)
	}).GetPendingTests()

	if err == nil {
		data.PendingTests = pending
	}

	return data, nil
}