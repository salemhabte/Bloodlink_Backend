package mocks

import (
	domain "bloodlink/Domain"
	"context"

	"github.com/stretchr/testify/mock"
)

// MockUserUseCase is a mock implementation of IUserUseCase
type MockUserUseCase struct {
	mock.Mock
}

func (m *MockUserUseCase) RegisterUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserUseCase) Login(ctx context.Context, email, password string) (string, string, string, string, bool, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.String(1), args.String(2), args.String(3), args.Bool(4), args.Error(5)
}

func (m *MockUserUseCase) RegisterDonor(ctx context.Context, req *domain.RegisterDonorRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockUserUseCase) SendOTP(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockUserUseCase) ResendOTP(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockUserUseCase) VerifyOTP(ctx context.Context, email, otp string) error {
	args := m.Called(ctx, email, otp)
	return args.Error(0)
}

func (m *MockUserUseCase) GetProfile(ctx context.Context, userID string) (*domain.ProfileResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProfileResponse), args.Error(1)
}

func (m *MockUserUseCase) UpdateProfile(ctx context.Context, profile *domain.UserProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockUserUseCase) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserUseCase) FilterDonors(ctx context.Context, filter domain.DonorFilter) (*domain.AllDonorsResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AllDonorsResponse), args.Error(1)
}

func (m *MockUserUseCase) ForgotPassword(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockUserUseCase) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	args := m.Called(ctx, email, otp, newPassword)
	return args.Error(0)
}

func (m *MockUserUseCase) UpdateDonorStatus(ctx context.Context, donorID, status string) error {
	args := m.Called(ctx, donorID, status)
	return args.Error(0)
}

func (m *MockUserUseCase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockUserUseCase) Logout(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserUseCase) GetUsersByRole(ctx context.Context, filter domain.UserFilter) ([]domain.UserResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserResponse), args.Error(1)
}

func (m *MockUserUseCase) GetAllProfiles(ctx context.Context, filter domain.ProfileFilter) ([]domain.UserProfile, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserProfile), args.Error(1)
}

func (m *MockUserUseCase) GetDonorIDByUserID(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockUserUseCase) GetEligibleDonors(ctx context.Context, query string) (*domain.EligibleDonorsResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EligibleDonorsResponse), args.Error(1)
}

func (m *MockUserUseCase) GetEligibleDonorByID(ctx context.Context, id string) (*domain.DonorResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DonorResponse), args.Error(1)
}

func (m *MockUserUseCase) NotifyEligibleDonors(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockUserUseCase) UpdateLocation(ctx context.Context, userID string, lat, lon float64) error {
	args := m.Called(ctx, userID, lat, lon)
	return args.Error(0)
}
