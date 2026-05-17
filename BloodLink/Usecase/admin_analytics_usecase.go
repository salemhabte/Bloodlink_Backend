package Usecase

import (
	"bloodlink/Domain"
	Interface "bloodlink/Domain/Interfaces"
)
type AdminAnalyticsUsecase struct {
	repo Interface.IAdminAnalyticsRepository
}

func NewAdminAnalyticsUsecase(r Interface.IAdminAnalyticsRepository) *AdminAnalyticsUsecase {
	return &AdminAnalyticsUsecase{repo: r}
}
func (u *AdminAnalyticsUsecase) GetDashboard() (*Domain.AdminDashboard, error) {

	d := &Domain.AdminDashboard{}

	// ================= DONORS =================
	donor, err := u.GetDonorSummary()
	if err != nil {
		return nil, err
	}

	d.TotalRegisteredDonors = donor.TotalRegisteredDonors
	d.DonorsWithDonationRecord = donor.DonorsWithDonationRecord
	d.ApprovedDonors = donor.ApprovedDonors
	d.TemporarilyRejectedDonors = donor.TemporarilyRejectedDonors

	// ================= SCREENING =================
	screening, err := u.GetScreeningSummary()
	if err != nil {
		return nil, err
	}

	d.ClearedDonors = screening.ClearedDonors
	d.TemporarilyDeferred = screening.TemporarilyDeferred
	d.PermanentlyDeferred = screening.PermanentlyDeferred
	d.ClearedPercent = screening.ClearedPercent
	d.TempDeferredPercent = screening.TempDeferredPercent
	d.PermanentDeferredPercent = screening.PermanentDeferredPercent

	// ================= COLLECTORS =================
	collector, err := u.GetCollectorSummary()
	if err != nil {
		return nil, err
	}

	d.TotalCollectors = collector.TotalCollectors
	d.TotalDonationRecords = collector.TotalDonationRecords
	d.DonationPerCollector = collector.DonationPerCollector

	// ================= LAB =================
	lab, err := u.GetLabSummary()
	if err != nil {
		return nil, err
	}

	d.TotalLabTechs = lab.TotalLabTechs
	d.TotalTestResults = lab.TotalTestResults
	d.LabCleared = lab.LabCleared
	d.LabTempDeferred = lab.LabTempDeferred
	d.LabPermDeferred = lab.LabPermDeferred
	d.LabClearedPercent = lab.LabClearedPercent
	d.LabTempPercent = lab.LabTempPercent
	d.LabPermPercent = lab.LabPermPercent
	d.TestsPerLabTech = lab.TestsPerLabTech

	// ================= INVENTORY =================
	inventory, err := u.GetInventorySummary()
	if err != nil {
		return nil, err
	}

	d.TotalBloodUnits = inventory.TotalBloodUnits
	d.BloodTypeStats = inventory.BloodTypeStats
	d.NearExpiryUnits = inventory.NearExpiryUnits

	// ================= HOSPITALS =================
	hospital, err := u.GetHospitalSummary()
	if err != nil {
		return nil, err
	}

	d.TotalHospitals = hospital.TotalHospitals
	d.ActiveContracts = hospital.ActiveContracts
	d.PendingHospitalRequests = hospital.PendingHospitalRequests
	d.ActiveEmergencies = hospital.ActiveEmergencies

	return d, nil
}
func (u *AdminAnalyticsUsecase) GetDonorSummary() (*Domain.DonorSummaryResponse, error) {

	total, withRecords, approved, rejected, err := u.repo.GetDonorStats()
	if err != nil {
		return nil, err
	}

	return &Domain.DonorSummaryResponse{
		TotalRegisteredDonors:     total,
		DonorsWithDonationRecord:  withRecords,
		ApprovedDonors:            approved,
		TemporarilyRejectedDonors: rejected,
	}, nil
}
func (u *AdminAnalyticsUsecase) GetScreeningSummary() (*Domain.ScreeningSummaryResponse, error) {

	cleared, temp, perm, err := u.repo.GetScreeningStats()
	if err != nil {
		return nil, err
	}

	total := cleared + temp + perm

	response := &Domain.ScreeningSummaryResponse{
		ClearedDonors:       cleared,
		TemporarilyDeferred: temp,
		PermanentlyDeferred: perm,
	}

	if total > 0 {
		response.ClearedPercent = float64(cleared) / float64(total) * 100
		response.TempDeferredPercent = float64(temp) / float64(total) * 100
		response.PermanentDeferredPercent = float64(perm) / float64(total) * 100
	}

	return response, nil
}
func (u *AdminAnalyticsUsecase) GetCollectorSummary() (*Domain.CollectorSummaryResponse, error) {

	tc, td, perCollector, err := u.repo.GetCollectorStats()
	if err != nil {
		return nil, err
	}

	return &Domain.CollectorSummaryResponse{
		TotalCollectors:      tc,
		TotalDonationRecords: td,
		DonationPerCollector: perCollector,
	}, nil
}
func (u *AdminAnalyticsUsecase) GetLabSummary() (*Domain.LabSummaryResponse, error) {

	// call repository
	tl, tt, lc, lt, lp, perLab, err := u.repo.GetLabStats()
	if err != nil {
		return nil, err
	}

	total := lc + lt + lp

	response := &Domain.LabSummaryResponse{
		TotalLabTechs:    tl,
		TotalTestResults: tt,
		LabCleared:       lc,
		LabTempDeferred:  lt,
		LabPermDeferred:  lp,
		TestsPerLabTech:  perLab, 
	}

	// ================================
	// CALCULATE PERCENTAGES
	// ================================
	if total > 0 {
		response.LabClearedPercent = float64(lc) / float64(total) * 100
		response.LabTempPercent = float64(lt) / float64(total) * 100
		response.LabPermPercent = float64(lp) / float64(total) * 100
	}

	return response, nil
}
func (u *AdminAnalyticsUsecase) GetInventorySummary() (*Domain.InventorySummaryResponse, error) {

	total, bloodTypes, nearExpiry, err := u.repo.GetInventoryStats()
	if err != nil {
		return nil, err
	}

	return &Domain.InventorySummaryResponse{
		TotalBloodUnits: total,
		BloodTypeStats:  bloodTypes,
		NearExpiryUnits: nearExpiry,
	}, nil
}

func (u *AdminAnalyticsUsecase) GetHospitalSummary() (*Domain.HospitalSummaryResponse, error) {
	total, active, pending, emergencies, err := u.repo.GetHospitalStats()
	if err != nil {
		return nil, err
	}

	return &Domain.HospitalSummaryResponse{
		TotalHospitals:          total,
		ActiveContracts:         active,
		PendingHospitalRequests: pending,
		ActiveEmergencies:       emergencies,
	}, nil
}