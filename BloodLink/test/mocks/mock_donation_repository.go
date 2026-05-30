package mocks

import (
	domain "bloodlink/Domain"

	"github.com/stretchr/testify/mock"
)

// MockDonationRepository mocks IDonationRepository
type MockDonationRepository struct {
	mock.Mock
}

func (m *MockDonationRepository) CreateDonation(record *domain.DonationRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockDonationRepository) SearchDonor(query string) (*domain.DonorResponse, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DonorResponse), args.Error(1)
}

func (m *MockDonationRepository) UpdateDonationStatus(donationID string, status string) error {
	args := m.Called(donationID, status)
	return args.Error(0)
}

func (m *MockDonationRepository) UpdateDonation(record *domain.DonationRecord) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockDonationRepository) GetDonationByID(id string) (*domain.DonationRecord, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DonationRecord), args.Error(1)
}

func (m *MockDonationRepository) GetLastDonationByDonor(donorID string) (*domain.DonationRecord, error) {
	args := m.Called(donorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DonationRecord), args.Error(1)
}

func (m *MockDonationRepository) UpdateDonorWeight(donorID string, weight float64) error {
	args := m.Called(donorID, weight)
	return args.Error(0)
}

func (m *MockDonationRepository) UpdateDonorOverallStatus(donorID string, status string) error {
	args := m.Called(donorID, status)
	return args.Error(0)
}

func (m *MockDonationRepository) GetPendingDonors() ([]domain.DonorResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonorResponse), args.Error(1)
}

func (m *MockDonationRepository) GetPendingDonorByID(donorID string) (*domain.DonorResponse, error) {
	args := m.Called(donorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DonorResponse), args.Error(1)
}

func (m *MockDonationRepository) SearchPendingDonor(query string) (*domain.DonorResponse, error) {
	args := m.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DonorResponse), args.Error(1)
}

func (m *MockDonationRepository) GetAllDonationsByDonor(donorID string) ([]domain.DonationRecord, error) {
	args := m.Called(donorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonationRecord), args.Error(1)
}

func (m *MockDonationRepository) GetDonorOverallStatus(donorID string) (string, error) {
	args := m.Called(donorID)
	return args.String(0), args.Error(1)
}

func (m *MockDonationRepository) GetDonationsByCollector(collectorID string) ([]domain.DonationRecord, error) {
	args := m.Called(collectorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonationRecord), args.Error(1)
}

func (m *MockDonationRepository) GetDonations(filter domain.DonationFilter) ([]domain.DonationRecord, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonationRecord), args.Error(1)
}
