package Domain

import "bloodlink/Domain"

type IHospitalRepository interface {
	CreateHospitalRequest(req *Domain.HospitalRequest) error
	CreateHospitalRequestAdmin(admin *Domain.HospitalRequestAdmin) error
	GetPendingRequests() ([]Domain.HospitalRequestResponse, error)
	GetHospitalRequestByID(requestID string) (*Domain.HospitalRequest, *Domain.HospitalRequestAdmin, error)
	UpdateHospitalRequestStatus(requestID string, status string) error

	CreateHospital(hospital *Domain.Hospital) error
	CreateHospitalAdmin(admin *Domain.HospitalAdmin) error
	GetHospitalAdminByUserID(userID string) (*Domain.HospitalAdmin, error)
	CreateContract(contract *Domain.HospitalContract) error

	GetContractByID(contractID string) (*Domain.HospitalContract, error)
	GetHospitalByID(hospitalID string) (*Domain.Hospital, error)
	UpdateContract(contract *Domain.HospitalContract) error

	CreateContractTemplate(template *Domain.ContractTemplate) error
	GetContractTemplates() ([]Domain.ContractTemplate, error)
	GetContractTemplateByID(templateID string) (*Domain.ContractTemplate, error)
	UpdateContractTemplate(template *Domain.ContractTemplate) error
	DeleteContractTemplate(templateID string) error
}

type IHospitalUsecase interface {
	SubmitRegistrationRequest(req *Domain.RegisterHospitalRequestDTO) error
	GetPendingRequests() ([]Domain.HospitalRequestResponse, error)
	ApproveRequest(requestID string, bloodBankAdminID string, payload *Domain.ApproveHospitalRequestDTO) error
	RejectRequest(requestID string) error
	HospitalSignContract(contractID string, req *Domain.SignContractRequestDTO, hospitalAdminID string) error
	AdminSignContract(contractID string, req *Domain.SignContractRequestDTO, bloodBankAdminID string) error
	RejectContract(contractID string, userID string, role string) error

	CreateContractTemplate(req *Domain.CreateTemplateRequestDTO, adminID string) error
	GetContractTemplates() ([]Domain.ContractTemplate, error)
	UpdateContractTemplate(templateID string, req *Domain.CreateTemplateRequestDTO) error
	DeleteContractTemplate(templateID string) error
}
