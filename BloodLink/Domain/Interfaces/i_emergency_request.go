package Domain

import "bloodlink/Domain"

type IEmergencyRequestRepository interface {
	Create(req *Domain.EmergencyRequest) error
	UpdateStatus(id string, status string) error
	GetByID(id string) (*Domain.EmergencyRequest, error)
	GetAll(filter Domain.EmergencyRequestFilter) ([]Domain.EmergencyRequest, error)
	GetActive() ([]Domain.EmergencyRequest, error)
	GetByRequestID(requestID string) (*Domain.EmergencyRequest, error)
	GetByLocation(location string) ([]Domain.EmergencyRequest, error)
	GetNearby(lat float64, lon float64, radiusKm float64, bloodType string) ([]Domain.EmergencyRequest, error)
	MarkCompletedEmergencies() error
}

type IEmergencyRequestUsecase interface {
	TriggerEmergency(requestID string, bloodType string, quantity int, urgencyLevel string, hospitalName string, location string, latitude float64, longitude float64) error
	PublishEmergency(id string) error
	RejectEmergency(id string) error
	CreateManualEmergency(req *Domain.CreateEmergencyRequestDTO) error
	GetAllEmergencies(filter Domain.EmergencyRequestFilter) (*Domain.EmergencyListResponse, error)
	GetPublishedEmergencies() ([]Domain.EmergencyRequest, error)
	GetEmergenciesForDonor(userID string) ([]Domain.EmergencyRequest, error)
	MarkCompletedEmergencies() error
}
