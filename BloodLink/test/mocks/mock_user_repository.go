package mocks

import (
	domain "bloodlink/Domain"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository mocks the IUserRepository interface used inside UserUseCaseBase
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) ActivateUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) CreateDonor(ctx context.Context, donor *domain.Donor) error {
	args := m.Called(ctx, donor)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) FilterDonors(ctx context.Context, filter domain.DonorFilter) ([]domain.DonorResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonorResponse), args.Error(1)
}

func (m *MockUserRepository) GetDonorStats(ctx context.Context) (*domain.AllDonorsResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AllDonorsResponse), args.Error(1)
}

func (m *MockUserRepository) SetOTP(ctx context.Context, email, otp string, expiresAt time.Time) error {
	args := m.Called(ctx, email, otp, expiresAt)
	return args.Error(0)
}

func (m *MockUserRepository) ResetPassword(ctx context.Context, email, hashedPassword string) error {
	args := m.Called(ctx, email, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateDonorStatus(ctx context.Context, donorID, status string) error {
	args := m.Called(ctx, donorID, status)
	return args.Error(0)
}

func (m *MockUserRepository) GetUsersByRole(ctx context.Context, filter domain.UserFilter) ([]domain.UserResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserResponse), args.Error(1)
}

func (m *MockUserRepository) UpdateRefreshToken(ctx context.Context, userID, refreshToken string) error {
	args := m.Called(ctx, userID, refreshToken)
	return args.Error(0)
}

func (m *MockUserRepository) GetDonorByUserID(ctx context.Context, userID string) (*domain.Donor, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Donor), args.Error(1)
}

func (m *MockUserRepository) GetDonorsByBloodTypeAndAddress(ctx context.Context, bloodType, address string) ([]domain.DonorResponse, error) {
	args := m.Called(ctx, bloodType, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonorResponse), args.Error(1)
}

func (m *MockUserRepository) GetDonorsNearby(ctx context.Context, bloodType string, lat, lon, radiusKm float64) ([]domain.DonorResponse, error) {
	args := m.Called(ctx, bloodType, lat, lon, radiusKm)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonorResponse), args.Error(1)
}

func (m *MockUserRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetEligibleDonors(ctx context.Context, query string) ([]domain.DonorResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonorResponse), args.Error(1)
}

func (m *MockUserRepository) GetEligibleDonorByID(ctx context.Context, id string) (*domain.DonorResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DonorResponse), args.Error(1)
}

func (m *MockUserRepository) GetDonorsBecomingEligibleToday(ctx context.Context) ([]domain.DonorResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DonorResponse), args.Error(1)
}
