package Domain

import (
	domain "bloodlink/Domain"
	"context"
)

type IHospitalRepository interface {
	Create(ctx context.Context, hospital *domain.Hospital) error
	GetByID(ctx context.Context, id string) (*domain.Hospital, error)
	Update(ctx context.Context, hospital *domain.Hospital) error
}

type IHospitalUsecase interface {
	RegisterHospital(ctx context.Context, req *domain.RegisterHospitalRequest) (*domain.Hospital, error)
	GetHospitalByID(ctx context.Context, id string) (*domain.Hospital, error)
	UpdateHospital(ctx context.Context, id string, req *domain.UpdateHospitalRequest) (*domain.Hospital, error)
}
