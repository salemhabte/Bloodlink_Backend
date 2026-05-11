package Domain

import "bloodlink/Domain"

type IBloodRequestRepository interface {
	CreateRequest(req *Domain.BloodRequest) error
	GetRequestsByHospital(filter Domain.BloodRequestFilter) ([]Domain.BloodRequestResponse, error)
	GetAllRequests(filter Domain.BloodRequestFilter) ([]Domain.BloodRequestResponse, error)
	GetRequestByID(requestID string) (*Domain.BloodRequest, error)
	UpdateRequestStatus(requestID string, status string, approvedAt *string) error
	UpdateRequestStatusWithDetails(requestID string, status string, approvedAt *string, notes string, fulfilledCount int, fulfilledVolumeMl int) error
	GetExpiredReservationRequests(cutoff string) ([]Domain.BloodRequest, error)
}

type IBloodRequestUsecase interface {
	CreateBloodRequest(req *Domain.CreateBloodRequestDTO, hospitalAdminID string) error
	GetHospitalRequests(filter Domain.BloodRequestFilter) ([]Domain.BloodRequestResponse, error)
	GetAllRequests(filter Domain.BloodRequestFilter) ([]Domain.BloodRequestResponse, error)
	ApproveRequest(requestID string) (*Domain.ApproveRequestResult, error)
	RejectRequest(requestID string) error
}
