package mocks

import (
	domain "bloodlink/Domain"

	"github.com/stretchr/testify/mock"
)

// MockBloodRequestRepository mocks IBloodRequestRepository
type MockBloodRequestRepository struct {
	mock.Mock
}

func (m *MockBloodRequestRepository) CreateRequest(req *domain.BloodRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockBloodRequestRepository) GetRequestsByHospital(filter domain.BloodRequestFilter) ([]domain.BloodRequestResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BloodRequestResponse), args.Error(1)
}

func (m *MockBloodRequestRepository) GetAllRequests(filter domain.BloodRequestFilter) ([]domain.BloodRequestResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BloodRequestResponse), args.Error(1)
}

func (m *MockBloodRequestRepository) GetRequestByID(requestID string) (*domain.BloodRequest, error) {
	args := m.Called(requestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.BloodRequest), args.Error(1)
}

func (m *MockBloodRequestRepository) UpdateRequestStatus(requestID string, status string, approvedAt *string) error {
	args := m.Called(requestID, status, approvedAt)
	return args.Error(0)
}

func (m *MockBloodRequestRepository) UpdateRequestStatusWithDetails(requestID string, status string, approvedAt *string, notes string, fulfilledCount int, fulfilledVolumeMl int) error {
	args := m.Called(requestID, status, approvedAt, notes, fulfilledCount, fulfilledVolumeMl)
	return args.Error(0)
}

func (m *MockBloodRequestRepository) GetExpiredReservationRequests(cutoff string) ([]domain.BloodRequest, error) {
	args := m.Called(cutoff)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BloodRequest), args.Error(1)
}
