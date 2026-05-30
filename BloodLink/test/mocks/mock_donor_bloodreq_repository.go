package mocks

import (
	domain "bloodlink/Domain"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockDonorBloodRequestRepository mocks IDonorBloodRequestRepository
type MockDonorBloodRequestRepository struct {
	mock.Mock
}

func (m *MockDonorBloodRequestRepository) Create(req *domain.DonorBloodRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockDonorBloodRequestRepository) GetAllAdmin(filter domain.DonorBloodRequestFilter) ([]domain.DonorBloodRequest, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonorBloodRequest), args.Error(1)
}

func (m *MockDonorBloodRequestRepository) GetByID(id string) (*domain.DonorBloodRequest, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DonorBloodRequest), args.Error(1)
}

func (m *MockDonorBloodRequestRepository) GetByDonorID(donorID string, filter domain.DonorBloodRequestFilter) ([]domain.DonorBloodRequest, error) {
	args := m.Called(donorID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonorBloodRequest), args.Error(1)
}

func (m *MockDonorBloodRequestRepository) UpdateStatus(id string, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockDonorBloodRequestRepository) UpdateStatusWithUnits(id string, status string, reservedUnits int) error {
	args := m.Called(id, status, reservedUnits)
	return args.Error(0)
}

func (m *MockDonorBloodRequestRepository) GetDonorIDByUserID(userID string) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockDonorBloodRequestRepository) GetDonorProfile(donorID string) (*domain.DonorProfile, error) {
	args := m.Called(donorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DonorProfile), args.Error(1)
}

func (m *MockDonorBloodRequestRepository) GetAvailableBloodUnits(bloodType string) ([]string, error) {
	args := m.Called(bloodType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDonorBloodRequestRepository) ReserveBloodUnits(requestID string, bloodType string, componentType string, requiredUnits int) (int, error) {
	args := m.Called(requestID, bloodType, componentType, requiredUnits)
	return args.Int(0), args.Error(1)
}

func (m *MockDonorBloodRequestRepository) MarkReservedAsUsed(requestID string) error {
	args := m.Called(requestID)
	return args.Error(0)
}

func (m *MockDonorBloodRequestRepository) ExpireStaleReservations() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDonorBloodRequestRepository) HasSuccessfulDonation(donorID string) (bool, error) {
	args := m.Called(donorID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDonorBloodRequestRepository) IsDonorInTop10(donorID string) (bool, error) {
	args := m.Called(donorID)
	return args.Bool(0), args.Error(1)
}

func (m *MockDonorBloodRequestRepository) GetLastRequestDateByDonor(donorID string) (time.Time, error) {
	args := m.Called(donorID)
	return args.Get(0).(time.Time), args.Error(1)
}
