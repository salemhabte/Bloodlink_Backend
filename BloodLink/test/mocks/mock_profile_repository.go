package mocks

import (
	domain "bloodlink/Domain"
	"context"

	"github.com/stretchr/testify/mock"
)

// MockProfileRepository mocks the IProfileRepository interface
type MockProfileRepository struct {
	mock.Mock
}

func (m *MockProfileRepository) CreateProfile(ctx context.Context, profile *domain.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockProfileRepository) GetProfileByUserID(ctx context.Context, userID string) (*domain.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserProfile), args.Error(1)
}

func (m *MockProfileRepository) UpdateProfile(ctx context.Context, profile *domain.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockProfileRepository) GetAllProfiles(ctx context.Context, filter domain.ProfileFilter) ([]domain.UserProfile, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserProfile), args.Error(1)
}
