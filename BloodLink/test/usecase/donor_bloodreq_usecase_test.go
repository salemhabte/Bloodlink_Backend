package usecase_test

import (
	"errors"
	"testing"
	"time"

	domain "bloodlink/Domain"
	"bloodlink/Usecase"
	"bloodlink/test/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// CreateRequest
// ─────────────────────────────────────────────

func TestCreateRequest_Success(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	userID := "user-123"
	donorID := "donor-456"
	donorProfile := &domain.DonorProfile{
		FullName:  "Abebe Kebede",
		Email:     "abebe@test.com",
		Phone:     "+251911111111",
		Address:   "Addis Ababa",
		BloodType: "O+",
	}

	repo.On("GetDonorIDByUserID", userID).Return(donorID, nil)
	repo.On("IsDonorInTop10", donorID).Return(true, nil)
	repo.On("GetLastRequestDateByDonor", donorID).Return(time.Time{}, nil) // zero = never requested
	repo.On("GetDonorProfile", donorID).Return(donorProfile, nil)
	repo.On("Create", mock.AnythingOfType("*Domain.DonorBloodRequest")).Return(nil)

	result, err := uc.CreateRequest(userID, 2, "PRBC", "Post-surgery", "St. Paul", "Addis Ababa", "0912345678")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "PENDING", result.Status)
	assert.Equal(t, donorID, result.DonorID)
	assert.Equal(t, 2, result.Units)
	assert.Equal(t, "PRBC", result.ComponentType)
	repo.AssertExpectations(t)
}

func TestCreateRequest_NotInTop10(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	userID := "user-123"
	donorID := "donor-456"

	repo.On("GetDonorIDByUserID", userID).Return(donorID, nil)
	repo.On("IsDonorInTop10", donorID).Return(false, nil)

	_, err := uc.CreateRequest(userID, 2, "PRBC", "reason", "hospital", "address", "phone")

	assert.EqualError(t, err, "Only top 10 leaderboard donors are eligible to request blood")
}

func TestCreateRequest_CooldownActive(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	userID := "user-123"
	donorID := "donor-456"
	recentDate := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago — within 90-day cooldown

	repo.On("GetDonorIDByUserID", userID).Return(donorID, nil)
	repo.On("IsDonorInTop10", donorID).Return(true, nil)
	repo.On("GetLastRequestDateByDonor", donorID).Return(recentDate, nil)

	_, err := uc.CreateRequest(userID, 2, "PRBC", "reason", "hospital", "address", "phone")

	assert.EqualError(t, err, "You can only request blood once every 3 months")
}

func TestCreateRequest_DonorIDNotFound(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	repo.On("GetDonorIDByUserID", "user-999").Return("", errors.New("donor not found"))

	_, err := uc.CreateRequest("user-999", 1, "PRBC", "reason", "hospital", "address", "phone")

	assert.Error(t, err)
}

// ─────────────────────────────────────────────
// ApproveRequest
// ─────────────────────────────────────────────

func TestApproveRequest_FullyApproved(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	requestID := "req-001"
	req := &domain.DonorBloodRequest{
		RequestID:     requestID,
		DonorID:       "donor-1",
		BloodType:     "O+",
		ComponentType: "PRBC",
		Units:         2,
		Status:        "PENDING",
	}
	updated := &domain.DonorBloodRequest{
		RequestID:     requestID,
		Status:        "APPROVED",
		ReservedUnits: 2,
	}

	repo.On("GetByID", requestID).Return(req, nil).Once()
	repo.On("ReserveBloodUnits", requestID, "O+", "PRBC", 2).Return(2, nil)
	repo.On("UpdateStatusWithUnits", requestID, "APPROVED", 2).Return(nil)
	repo.On("GetByID", requestID).Return(updated, nil).Once()

	result, message, err := uc.ApproveRequest(requestID)

	assert.NoError(t, err)
	assert.Equal(t, "fully approved", message)
	assert.Equal(t, "APPROVED", result.Status)
}

func TestApproveRequest_PartiallyApproved(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	requestID := "req-002"
	req := &domain.DonorBloodRequest{
		RequestID:     requestID,
		BloodType:     "A+",
		ComponentType: "PLATELETS",
		Units:         3,
		Status:        "PENDING",
	}
	updated := &domain.DonorBloodRequest{
		RequestID:     requestID,
		Status:        "PARTIALLY APPROVED",
		ReservedUnits: 1,
	}

	repo.On("GetByID", requestID).Return(req, nil).Once()
	repo.On("ReserveBloodUnits", requestID, "A+", "PLATELETS", 3).Return(1, nil) // only 1 available
	repo.On("UpdateStatusWithUnits", requestID, "PARTIALLY APPROVED", 1).Return(nil)
	repo.On("GetByID", requestID).Return(updated, nil).Once()

	result, message, err := uc.ApproveRequest(requestID)

	assert.NoError(t, err)
	assert.Equal(t, "partially approved", message)
	assert.Equal(t, "PARTIALLY APPROVED", result.Status)
}

func TestApproveRequest_NoBloodAvailable(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	requestID := "req-003"
	req := &domain.DonorBloodRequest{
		RequestID:     requestID,
		BloodType:     "AB-",
		ComponentType: "PLASMA",
		Units:         2,
		Status:        "PENDING",
	}
	updated := &domain.DonorBloodRequest{
		RequestID: requestID,
		Status:    "REJECTED",
	}

	repo.On("GetByID", requestID).Return(req, nil).Once()
	repo.On("ReserveBloodUnits", requestID, "AB-", "PLASMA", 2).Return(0, nil)
	repo.On("UpdateStatusWithUnits", requestID, "REJECTED", 0).Return(nil)
	repo.On("GetByID", requestID).Return(updated, nil).Once()

	result, message, err := uc.ApproveRequest(requestID)

	assert.NoError(t, err)
	assert.Equal(t, "no enough blood in the inventory", message)
	assert.Equal(t, "REJECTED", result.Status)
}

func TestApproveRequest_AlreadyProcessed(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	requestID := "req-004"
	req := &domain.DonorBloodRequest{
		RequestID: requestID,
		Status:    "APPROVED", // already processed
	}

	repo.On("GetByID", requestID).Return(req, nil)

	_, _, err := uc.ApproveRequest(requestID)

	assert.EqualError(t, err, "request already processed")
}

// ─────────────────────────────────────────────
// RejectRequest
// ─────────────────────────────────────────────

func TestRejectRequest_Success(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	requestID := "req-005"
	req := &domain.DonorBloodRequest{RequestID: requestID, Status: "PENDING"}
	updated := &domain.DonorBloodRequest{RequestID: requestID, Status: "REJECTED"}

	repo.On("GetByID", requestID).Return(req, nil).Once()
	repo.On("UpdateStatus", requestID, "REJECTED").Return(nil)
	repo.On("GetByID", requestID).Return(updated, nil).Once()

	result, err := uc.RejectRequest(requestID)

	assert.NoError(t, err)
	assert.Equal(t, "REJECTED", result.Status)
}

func TestRejectRequest_AlreadyFulfilled(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	requestID := "req-006"
	req := &domain.DonorBloodRequest{RequestID: requestID, Status: "FULFILLED"}

	repo.On("GetByID", requestID).Return(req, nil)

	_, err := uc.RejectRequest(requestID)

	assert.EqualError(t, err, "cannot reject a fulfilled request")
}

// ─────────────────────────────────────────────
// FulfillRequest
// ─────────────────────────────────────────────

func TestFulfillRequest_ApprovedToFulfilled(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	requestID := "req-007"
	req := &domain.DonorBloodRequest{RequestID: requestID, Status: "APPROVED"}
	updated := &domain.DonorBloodRequest{RequestID: requestID, Status: "FULFILLED"}

	repo.On("GetByID", requestID).Return(req, nil).Once()
	repo.On("MarkReservedAsUsed", requestID).Return(nil)
	repo.On("UpdateStatus", requestID, "FULFILLED").Return(nil)
	repo.On("GetByID", requestID).Return(updated, nil).Once()

	result, err := uc.FulfillRequest(requestID)

	assert.NoError(t, err)
	assert.Equal(t, "FULFILLED", result.Status)
}

func TestFulfillRequest_PartiallyApprovedToPartiallyFulfilled(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	requestID := "req-008"
	req := &domain.DonorBloodRequest{RequestID: requestID, Status: "PARTIALLY APPROVED"}
	updated := &domain.DonorBloodRequest{RequestID: requestID, Status: "PARTIALLY FULFILLED"}

	repo.On("GetByID", requestID).Return(req, nil).Once()
	repo.On("MarkReservedAsUsed", requestID).Return(nil)
	repo.On("UpdateStatus", requestID, "PARTIALLY FULFILLED").Return(nil)
	repo.On("GetByID", requestID).Return(updated, nil).Once()

	result, err := uc.FulfillRequest(requestID)

	assert.NoError(t, err)
	assert.Equal(t, "PARTIALLY FULFILLED", result.Status)
}

func TestFulfillRequest_PendingBlocked(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	requestID := "req-009"
	req := &domain.DonorBloodRequest{RequestID: requestID, Status: "PENDING"}

	repo.On("GetByID", requestID).Return(req, nil)

	_, err := uc.FulfillRequest(requestID)

	assert.EqualError(t, err, "request has not been approved yet")
}

// ─────────────────────────────────────────────
// GetMyRequests
// ─────────────────────────────────────────────

func TestGetMyRequests_ReturnsFilteredList(t *testing.T) {
	repo := &mocks.MockDonorBloodRequestRepository{}
	uc := Usecase.NewDonorBloodRequestUsecase(repo)

	userID := "user-123"
	donorID := "donor-456"
	filter := domain.DonorBloodRequestFilter{Status: "PENDING"}

	requests := []domain.DonorBloodRequest{
		{RequestID: "r1", Status: "PENDING"},
		{RequestID: "r2", Status: "PENDING"},
	}

	repo.On("GetDonorIDByUserID", userID).Return(donorID, nil)
	repo.On("GetByDonorID", donorID, filter).Return(requests, nil)

	result, err := uc.GetMyRequests(userID, filter)

	assert.NoError(t, err)
	assert.Len(t, result.Requests, 2)
	assert.Equal(t, 2, result.Analytics.TotalPending)
}
