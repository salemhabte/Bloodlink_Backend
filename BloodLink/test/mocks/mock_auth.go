package mocks

import (
	domain "bloodlink/Domain"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockAuthentication mocks IAuthentication
type MockAuthentication struct {
	mock.Mock
}

func (m *MockAuthentication) ParseTokenToClaim(token string) (*domain.UserClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserClaims), args.Error(1)
}

func (m *MockAuthentication) GenerateToken(claims *domain.UserClaims, tokenType string) (string, error) {
	args := m.Called(claims, tokenType)
	return args.String(0), args.Error(1)
}

// MockUserValidation mocks IUserValidation
type MockUserValidation struct {
	mock.Mock
}

func (m *MockUserValidation) IsValidEmail(email string) bool {
	args := m.Called(email)
	return args.Bool(0)
}

func (m *MockUserValidation) IsStrongPassword(password string) bool {
	args := m.Called(password)
	return args.Bool(0)
}

func (m *MockUserValidation) Hashpassword(password string) string {
	args := m.Called(password)
	return args.String(0)
}

func (m *MockUserValidation) ComparePassword(userPassword, password string) error {
	args := m.Called(userPassword, password)
	return args.Error(0)
}

func (m *MockUserValidation) IsValidPhone(phone string) bool {
	args := m.Called(phone)
	return args.Bool(0)
}

// MockNotificationUsecase mocks INotificationUsecase
type MockNotificationUsecase struct {
	mock.Mock
}

func (m *MockNotificationUsecase) SendNotification(userID, notifType, title, message string) error {
	args := m.Called(userID, notifType, title, message)
	return args.Error(0)
}

func (m *MockNotificationUsecase) SendToRole(role, notifType, title, message string) error {
	args := m.Called(role, notifType, title, message)
	return args.Error(0)
}

func (m *MockNotificationUsecase) SendToHospital(hospitalID, notifType, title, message string) error {
	args := m.Called(hospitalID, notifType, title, message)
	return args.Error(0)
}

func (m *MockNotificationUsecase) SendToDonor(donorID, notifType, title, message string) error {
	args := m.Called(donorID, notifType, title, message)
	return args.Error(0)
}

func (m *MockNotificationUsecase) GetMyNotifications(userID string) ([]domain.Notification, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Notification), args.Error(1)
}

func (m *MockNotificationUsecase) MarkAsRead(notificationID string, userID string) error {
	args := m.Called(notificationID, userID)
	return args.Error(0)
}

func (m *MockNotificationUsecase) MarkAllAsRead(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// MockHospitalRepository mocks IHospitalRepository (partial — only methods used in tests)
type MockHospitalRepository struct {
	mock.Mock
}

func (m *MockHospitalRepository) GetHospitalAdminByUserID(userID string) (*domain.HospitalAdmin, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HospitalAdmin), args.Error(1)
}

func (m *MockHospitalRepository) IsAdminEmailPending(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockHospitalRepository) GetContractsByHospitalID(hospitalID string) ([]domain.HospitalContract, error) {
	args := m.Called(hospitalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.HospitalContract), args.Error(1)
}

func (m *MockHospitalRepository) GetHospitalByID(hospitalID string) (*domain.Hospital, error) {
	args := m.Called(hospitalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Hospital), args.Error(1)
}

// satisfy the full interface — remaining methods are no-ops for tests that don't need them
func (m *MockHospitalRepository) CreateHospitalRequest(req *domain.HospitalRequest) error {
	return nil
}
func (m *MockHospitalRepository) CreateHospitalRequestAdmin(admin *domain.HospitalRequestAdmin) error {
	return nil
}
func (m *MockHospitalRepository) CreateHospitalRegistrationRequest(req *domain.HospitalRequest, admin *domain.HospitalRequestAdmin) error {
	return nil
}
func (m *MockHospitalRepository) ApproveHospitalRegistration(hospital *domain.Hospital, user *domain.User, admin *domain.HospitalAdmin, contract *domain.HospitalContract, requestID string) error {
	return nil
}
func (m *MockHospitalRepository) GetPendingRequests(filter domain.HospitalRequestFilter) ([]domain.HospitalRequestResponse, error) {
	return nil, nil
}
func (m *MockHospitalRepository) GetHospitalRequestByID(requestID string) (*domain.HospitalRequest, *domain.HospitalRequestAdmin, error) {
	return nil, nil, nil
}
func (m *MockHospitalRepository) UpdateHospitalRequestStatus(requestID string, status string) error {
	return nil
}
func (m *MockHospitalRepository) CreateHospital(hospital *domain.Hospital) error { return nil }
func (m *MockHospitalRepository) CreateHospitalAdmin(admin *domain.HospitalAdmin) error {
	return nil
}
func (m *MockHospitalRepository) CreateContract(contract *domain.HospitalContract) error {
	return nil
}
func (m *MockHospitalRepository) GetContractByID(contractID string) (*domain.HospitalContract, error) {
	return nil, nil
}
func (m *MockHospitalRepository) UpdateContract(contract *domain.HospitalContract) error {
	return nil
}
func (m *MockHospitalRepository) CreateContractTemplate(template *domain.ContractTemplate) error {
	return nil
}
func (m *MockHospitalRepository) GetContractTemplates() ([]domain.ContractTemplate, error) {
	return nil, nil
}
func (m *MockHospitalRepository) GetContractTemplateByID(templateID string) (*domain.ContractTemplate, error) {
	return nil, nil
}
func (m *MockHospitalRepository) UpdateContractTemplate(template *domain.ContractTemplate) error {
	return nil
}
func (m *MockHospitalRepository) DeleteContractTemplate(templateID string) error { return nil }
func (m *MockHospitalRepository) GetSignedContracts(status string) ([]domain.HospitalContractResponse, error) {
	return nil, nil
}
func (m *MockHospitalRepository) GetHospitalDashboard(hospitalID string) (*domain.HospitalDashboard, error) {
	return nil, nil
}
func (m *MockHospitalRepository) GetHospitalByPhone(phone string) (*domain.Hospital, error) {
	return nil, nil
}
func (m *MockHospitalRepository) GetHospitalByName(name string) (*domain.Hospital, error) {
	return nil, nil
}
func (m *MockHospitalRepository) GetAllHospitals() ([]domain.Hospital, error) { return nil, nil }
func (m *MockHospitalRepository) IsPhoneRegisteredOrPending(phone string) (bool, error) {
	return false, nil
}

// MockBloodInventoryRepository mocks IBloodInventoryRepository (partial)
type MockBloodInventoryRepository struct {
	mock.Mock
}

func (m *MockBloodInventoryRepository) ReserveUnitsForHospital(bloodType string, componentType string, quantity int, hospitalID string, requestID string) ([]domain.BloodUnit, error) {
	args := m.Called(bloodType, componentType, quantity, hospitalID, requestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BloodUnit), args.Error(1)
}

func (m *MockBloodInventoryRepository) GetReservedUnitsByRequestID(requestID string) ([]domain.BloodUnit, error) {
	args := m.Called(requestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BloodUnit), args.Error(1)
}

func (m *MockBloodInventoryRepository) MarkUnitAsUsed(unitID string) error {
	args := m.Called(unitID)
	return args.Error(0)
}

func (m *MockBloodInventoryRepository) CountAvailableUnitsByBloodType(bloodType string) (int, error) {
	args := m.Called(bloodType)
	return args.Int(0), args.Error(1)
}

// satisfy remaining interface methods
func (m *MockBloodInventoryRepository) GetAllBloodUnits(filter domain.BloodUnitFilter) ([]domain.BloodUnit, error) {
	return nil, nil
}
func (m *MockBloodInventoryRepository) GetBloodUnitByID(id string) (*domain.BloodUnit, error) {
	return nil, nil
}
func (m *MockBloodInventoryRepository) UpdateBloodUnitStatus(id string, status string) error {
	return nil
}
func (m *MockBloodInventoryRepository) DeleteBloodUnitByID(id string) error { return nil }
func (m *MockBloodInventoryRepository) GetFullBloodUnitDetails(id string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockBloodInventoryRepository) FilterBloodUnits(filter domain.BloodUnitFilter) ([]domain.BloodUnit, error) {
	return nil, nil
}
func (m *MockBloodInventoryRepository) MarkExpiredUnits() error { return nil }
func (m *MockBloodInventoryRepository) ConsumeUnits(bloodType string, quantity int) error {
	return nil
}
func (m *MockBloodInventoryRepository) ExpireStaleReservations(cutoff time.Time) ([]string, error) {
	return nil, nil
}
func (m *MockBloodInventoryRepository) GetReservedUnitsByHospitalID(hospitalID string) ([]domain.BloodUnit, error) {
	return nil, nil
}
func (m *MockBloodInventoryRepository) DeleteWithAudit(unitID string) error { return nil }
func (m *MockBloodInventoryRepository) ConvertPlasmaToCryo(plasmaUnitID string, cryo *domain.BloodUnit, cryoPoor *domain.BloodUnit) error {
	return nil
}
func (m *MockBloodInventoryRepository) IsSlotOccupied(location, rack, shelf, position string) (bool, error) {
	return false, nil
}
func (m *MockBloodInventoryRepository) GetOccupiedSlotCount(location, rack, shelf string) (int, error) {
	return 0, nil
}
