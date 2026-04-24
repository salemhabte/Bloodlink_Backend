package Domain

import "bloodlink/Domain"
type ICollectorAnalyticsRepository interface {

	GetCollectorKPI(collectorID string) (*Domain.CollectorKPI, error)

	GetDonorInsights(collectorID string) (*Domain.CollectorDonorInsights, error)

	GetTodayStats(collectorID string) (*Domain.CollectorTodayStats, error)
}