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

func buildBloodRequestUsecase(
	repo *mocks.MockBloodRequestRepository,
	hospitalRepo *mocks.MockHospitalRepository,
	inventoryRepo *mocks.MockBloodInventoryRepository,
	notif *mocks.MockNotificationUsecase,
) *bloodRequestUsecaseWrapper {
	// We need a mock emergency usecase too
	emergencyUC := &mockEmergencyUsecase{}
	uc := Usecase.NewBloodRequestUsecase(repo, hospitalRepo, inventoryRepo, emergencyUC, notif)
	return &bloodRequestUsecaseWrapper{uc: uc}
}

type bloodRequestUsecaseWrapper struct {
	uc interface {
		ApproveRequest(requestID string) (*domain.ApproveRequestResult, error)
		RejectRequest(requestID string) error
		GetAllRequests(filter domain.BloodRequestFilter) (*domain.BloodRequestListResponse, error)
		GetHospitalRequests(filter domain.BloodRequestFilter) (*domain.BloodRequestListResponse, error)
		MarkRequestUnitsAsUsed(requestID string) error
		CreateBloodRequest(req *domain.CreateBloodRequestBatchDTO, hospitalAdminID string) error
	}
}

// minimal mock for IEmergencyRequestUsecase — only TriggerEmergency is called in blood_request_usecase
type mockEmergencyUsecase struct{}

func (m *mockEmergencyUsecase) TriggerEmergency(requestID string, bloodType string, quantity int, urgencyLevel string, hospitalName string, location string, latitude float64, longitude float64) error {
	return nil
}
func (m *mockEmergencyUsecase) PublishEmergency(id string) error  { return nil }
func (m *mockEmergencyUsecase) RejectEmergency(id string) error   { return nil }
func (m *mockEmergencyUsecase) CreateManualEmergency(reqs []domain.CreateEmergencyRequestDTO) error {
	return nil
}
func (m *mockEmergencyUsecase) GetAllEmergencies(filter domain.EmergencyRequestFilter) (*domain.EmergencyListResponse, error) {
	return nil, nil
}
func (m *mockEmergencyUsecase) GetPublishedEmergencies() ([]domain.EmergencyRequest, error) {
	return nil, nil
}
func (m *mockEmergencyUsecase) GetEmergenciesForDonor(userID string) ([]domain.EmergencyRequest, error) {
	return nil, nil
}
func (m *mockEmergencyUsecase) MarkCompletedEmergencies() error { return nil }

// ─────────────────────────────────────────────
// ApproveRequest
// ─────────────────────────────────────────────

func TestBloodRequest_ApproveRequest_FullyFulfilled(t *testing.T) {
	repo := &mocks.MockBloodRequestRepository{}
	hospitalRepo := &mocks.MockHospitalRepository{}
	inventoryRepo := &mocks.MockBloodInventoryRepository{}
	notif := &mocks.MockNotificationUsecase{}

	w := buildBloodRequestUsecase(repo, hospitalRepo, inventoryRepo, notif)

	requestID := "br-001"
	br := &domain.BloodRequest{
		RequestID:  requestID,
		HospitalID: "hosp-1",
		BloodType:  "O+",
		Component:  "PRBC",
		Quantity:   2,
		Status:     domain.BloodRequestStatusPending,
	}

	reservedUnits := []domain.BloodUnit{
		{BloodUnitID: "u1", BloodType: "O+", QuantityML: 300, ExpirationDate: time.Now().Add(30 * 24 * time.Hour)},
		{BloodUnitID: "u2", BloodType: "O+", QuantityML: 300, ExpirationDate: time.Now().Add(20 * 24 * time.Hour)},
	}

	repo.On("GetRequestByID", requestID).Return(br, nil)
	inventoryRepo.On("ReserveUnitsForHospital", "O+", "PRBC", 2, "hosp-1", requestID).Return(reservedUnits, nil)
	repo.On("UpdateRequestStatusWithDetails", requestID, domain.BloodRequestStatusFulfilled, mock.Anything, mock.Anything, 2, 600).Return(nil)
	notif.On("SendToHospital", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	result, err := w.uc.ApproveRequest(requestID)

	assert.NoError(t, err)
	assert.Equal(t, domain.BloodRequestStatusFulfilled, result.Status)
	assert.Equal(t, 2, result.FulfilledCount)
	assert.Equal(t, 600, result.TotalQuantityML)
}

func TestBloodRequest_ApproveRequest_NoInventory_AutoRejects(t *testing.T) {
	repo := &mocks.MockBloodRequestRepository{}
	hospitalRepo := &mocks.MockHospitalRepository{}
	inventoryRepo := &mocks.MockBloodInventoryRepository{}
	notif := &mocks.MockNotificationUsecase{}

	w := buildBloodRequestUsecase(repo, hospitalRepo, inventoryRepo, notif)

	requestID := "br-002"
	br := &domain.BloodRequest{
		RequestID:  requestID,
		HospitalID: "hosp-1",
		BloodType:  "AB-",
		Component:  "PLASMA",
		Quantity:   3,
		Status:     domain.BloodRequestStatusPending,
	}

	repo.On("GetRequestByID", requestID).Return(br, nil)
	inventoryRepo.On("ReserveUnitsForHospital", "AB-", "PLASMA", 3, "hosp-1", requestID).Return([]domain.BloodUnit{}, nil)
	repo.On("UpdateRequestStatusWithDetails", requestID, domain.BloodRequestStatusRejected, mock.Anything, mock.Anything, 0, 0).Return(nil)
	notif.On("SendToHospital", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	result, err := w.uc.ApproveRequest(requestID)

	assert.NoError(t, err)
	assert.Equal(t, domain.BloodRequestStatusRejected, result.Status)
	assert.Equal(t, 0, result.FulfilledCount)
}

func TestBloodRequest_ApproveRequest_AlreadyProcessed(t *testing.T) {
	repo := &mocks.MockBloodRequestRepository{}
	hospitalRepo := &mocks.MockHospitalRepository{}
	inventoryRepo := &mocks.MockBloodInventoryRepository{}
	notif := &mocks.MockNotificationUsecase{}

	w := buildBloodRequestUsecase(repo, hospitalRepo, inventoryRepo, notif)

	requestID := "br-003"
	br := &domain.BloodRequest{
		RequestID: requestID,
		Status:    domain.BloodRequestStatusFulfilled,
	}

	repo.On("GetRequestByID", requestID).Return(br, nil)

	_, err := w.uc.ApproveRequest(requestID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot approve")
}

// ─────────────────────────────────────────────
// RejectRequest
// ─────────────────────────────────────────────

func TestBloodRequest_RejectRequest_Success(t *testing.T) {
	repo := &mocks.MockBloodRequestRepository{}
	hospitalRepo := &mocks.MockHospitalRepository{}
	inventoryRepo := &mocks.MockBloodInventoryRepository{}
	notif := &mocks.MockNotificationUsecase{}

	w := buildBloodRequestUsecase(repo, hospitalRepo, inventoryRepo, notif)

	requestID := "br-004"
	br := &domain.BloodRequest{
		RequestID:  requestID,
		HospitalID: "hosp-1",
		BloodType:  "B+",
		Quantity:   1,
		Status:     domain.BloodRequestStatusPending,
	}

	repo.On("GetRequestByID", requestID).Return(br, nil)
	repo.On("UpdateRequestStatus", requestID, domain.BloodRequestStatusRejected, (*string)(nil)).Return(nil)
	notif.On("SendToHospital", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	err := w.uc.RejectRequest(requestID)

	assert.NoError(t, err)
}

func TestBloodRequest_RejectRequest_NotPending(t *testing.T) {
	repo := &mocks.MockBloodRequestRepository{}
	hospitalRepo := &mocks.MockHospitalRepository{}
	inventoryRepo := &mocks.MockBloodInventoryRepository{}
	notif := &mocks.MockNotificationUsecase{}

	w := buildBloodRequestUsecase(repo, hospitalRepo, inventoryRepo, notif)

	requestID := "br-005"
	br := &domain.BloodRequest{
		RequestID: requestID,
		Status:    domain.BloodRequestStatusFulfilled,
	}

	repo.On("GetRequestByID", requestID).Return(br, nil)

	err := w.uc.RejectRequest(requestID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot reject")
}

// ─────────────────────────────────────────────
// MarkRequestUnitsAsUsed
// ─────────────────────────────────────────────

func TestBloodRequest_MarkUnitsAsUsed_Success(t *testing.T) {
	repo := &mocks.MockBloodRequestRepository{}
	hospitalRepo := &mocks.MockHospitalRepository{}
	inventoryRepo := &mocks.MockBloodInventoryRepository{}
	notif := &mocks.MockNotificationUsecase{}

	w := buildBloodRequestUsecase(repo, hospitalRepo, inventoryRepo, notif)

	requestID := "br-006"
	br := &domain.BloodRequest{
		RequestID: requestID,
		Status:    domain.BloodRequestStatusFulfilled,
	}
	units := []domain.BloodUnit{
		{BloodUnitID: "u1"},
		{BloodUnitID: "u2"},
	}

	repo.On("GetRequestByID", requestID).Return(br, nil)
	inventoryRepo.On("GetReservedUnitsByRequestID", requestID).Return(units, nil)
	inventoryRepo.On("MarkUnitAsUsed", "u1").Return(nil)
	inventoryRepo.On("MarkUnitAsUsed", "u2").Return(nil)

	err := w.uc.MarkRequestUnitsAsUsed(requestID)

	assert.NoError(t, err)
	inventoryRepo.AssertCalled(t, "MarkUnitAsUsed", "u1")
	inventoryRepo.AssertCalled(t, "MarkUnitAsUsed", "u2")
}

func TestBloodRequest_MarkUnitsAsUsed_WrongStatus(t *testing.T) {
	repo := &mocks.MockBloodRequestRepository{}
	hospitalRepo := &mocks.MockHospitalRepository{}
	inventoryRepo := &mocks.MockBloodInventoryRepository{}
	notif := &mocks.MockNotificationUsecase{}

	w := buildBloodRequestUsecase(repo, hospitalRepo, inventoryRepo, notif)

	requestID := "br-007"
	br := &domain.BloodRequest{
		RequestID: requestID,
		Status:    domain.BloodRequestStatusPending,
	}

	repo.On("GetRequestByID", requestID).Return(br, nil)

	err := w.uc.MarkRequestUnitsAsUsed(requestID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot mark as used")
}

// ─────────────────────────────────────────────
// GetAllRequests
// ─────────────────────────────────────────────

func TestBloodRequest_GetAllRequests_WithAnalytics(t *testing.T) {
	repo := &mocks.MockBloodRequestRepository{}
	hospitalRepo := &mocks.MockHospitalRepository{}
	inventoryRepo := &mocks.MockBloodInventoryRepository{}
	notif := &mocks.MockNotificationUsecase{}

	w := buildBloodRequestUsecase(repo, hospitalRepo, inventoryRepo, notif)

	filter := domain.BloodRequestFilter{}
	requests := []domain.BloodRequestResponse{
		{RequestID: "r1", Status: domain.BloodRequestStatusPending},
		{RequestID: "r2", Status: domain.BloodRequestStatusFulfilled},
		{RequestID: "r3", Status: domain.BloodRequestStatusRejected},
	}

	repo.On("GetAllRequests", filter).Return(requests, nil)

	result, err := w.uc.GetAllRequests(filter)

	assert.NoError(t, err)
	assert.Len(t, result.Requests, 3)
	assert.Equal(t, 3, result.Analytics.TotalRequests)
	assert.Equal(t, 1, result.Analytics.TotalPending)
	assert.Equal(t, 1, result.Analytics.TotalFulfilled)
	assert.Equal(t, 1, result.Analytics.TotalCancelled)
}

func TestBloodRequest_GetAllRequests_RepoError(t *testing.T) {
	repo := &mocks.MockBloodRequestRepository{}
	hospitalRepo := &mocks.MockHospitalRepository{}
	inventoryRepo := &mocks.MockBloodInventoryRepository{}
	notif := &mocks.MockNotificationUsecase{}

	w := buildBloodRequestUsecase(repo, hospitalRepo, inventoryRepo, notif)

	filter := domain.BloodRequestFilter{}
	repo.On("GetAllRequests", filter).Return(nil, errors.New("db error"))

	_, err := w.uc.GetAllRequests(filter)

	assert.EqualError(t, err, "db error")
}
