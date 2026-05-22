package Domain

import "bloodlink/Domain"

type ILabAnalyticsRepository interface {
	GetLabDashboard(labID string) (*Domain.LabDashboard, error)
	GetPendingTests() (int, error)
}