package Domain

import "bloodlink/Domain"

type IEmergencyRequestRepository interface {
	Create(req *Domain.EmergencyRequest) error
	UpdateStatus(id string, status string) error
	GetByID(id string) (*Domain.EmergencyRequest, error)
	GetAll() ([]Domain.EmergencyRequest, error)
	GetActive() ([]Domain.EmergencyRequest, error)
	GetByRequestID(requestID string) (*Domain.EmergencyRequest, error)
	GetByLocation(location string) ([]Domain.EmergencyRequest, error)
}

type IEmergencyRequestUsecase interface {
	TriggerEmergency(requestID string, bloodType string, quantity int, urgencyLevel string, hospitalName string, location string) error
	PublishEmergency(id string) error
	RejectEmergency(id string) error
	CreateManualEmergency(req *Domain.CreateEmergencyRequestDTO) error
	GetAllEmergencies() ([]Domain.EmergencyRequest, error)
	GetPublishedEmergencies() ([]Domain.EmergencyRequest, error)
	GetEmergenciesForDonor(userID string) ([]Domain.EmergencyRequest, error)
	GetEmergencyByID(id string) (*Domain.EmergencyRequest, error)
}
