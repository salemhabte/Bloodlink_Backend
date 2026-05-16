package Domain

import "bloodlink/Domain"

type IAdminAnalyticsRepository interface {
	GetDonorStats() (donors, withRecords, approved, rejected int, err error)
	GetScreeningStats() (cleared, temp, perm int, err error)
	GetCollectorStats() (totalCollectors int, totalDonations int, perCollector []Domain.CollectorDonationStats, err error)
	GetLabStats() (totalLabs, totalTests, cleared, temp, perm int, perLab []Domain.LabTestStats, err error)
	GetInventoryStats() (totalUnits int, bloodTypes []Domain.BloodTypeStat, nearExpiry int, err error)
	GetHospitalStats() (totalHospitals, activeContracts, pendingRequests, activeEmergencies int, err error)
}