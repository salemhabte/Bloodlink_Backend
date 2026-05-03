package Domain

import "bloodlink/Domain"

type IBloodRequestRepository interface {
	CreateRequest(req *Domain.BloodRequest) error
	GetRequestsByHospital(filter Domain.BloodRequestFilter) ([]Domain.BloodRequestResponse, error)
	GetAllRequests() ([]Domain.BloodRequestResponse, error)
	GetRequestByID(requestID string) (*Domain.BloodRequest, error)
	UpdateRequestStatus(requestID string, status string, approvedAt *string) error
}

type IBloodRequestUsecase interface {
	CreateBloodRequest(req *Domain.CreateBloodRequestDTO, hospitalAdminID string) error
	GetHospitalRequests(filter Domain.BloodRequestFilter) ([]Domain.BloodRequestResponse, error)
	GetAllRequests() ([]Domain.BloodRequestResponse, error)
	UpdateStatus(requestID string, req *Domain.UpdateBloodRequestStatusDTO) error
}
