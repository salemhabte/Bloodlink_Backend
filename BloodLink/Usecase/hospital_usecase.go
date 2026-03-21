package Usecase

import (
	domain "bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type HospitalUsecase struct {
	repo domainInterface.IHospitalRepository
}

func NewHospitalUsecase(repo domainInterface.IHospitalRepository) domainInterface.IHospitalUsecase {
	return &HospitalUsecase{repo: repo}
}

func (u *HospitalUsecase) RegisterHospital(ctx context.Context, req *domain.RegisterHospitalRequest) (*domain.Hospital, error) {
	hospital := &domain.Hospital{
		HospitalID:         uuid.New().String(),
		HospitalName:       req.HospitalName,
		Address:            req.Address,
		City:               req.City,
		Phone:              req.Phone,
		ContactPersonName:  req.ContactPersonName,
		ContactPersonPhone: req.ContactPersonPhone,
		Status:             "PENDING",
		CreatedAt:          time.Now(),
	}

	if err := u.repo.Create(ctx, hospital); err != nil {
		return nil, err
	}

	return hospital, nil
}

func (u *HospitalUsecase) GetHospitalByID(ctx context.Context, id string) (*domain.Hospital, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *HospitalUsecase) UpdateHospital(ctx context.Context, id string, req *domain.UpdateHospitalRequest) (*domain.Hospital, error) {
	hospital, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.HospitalName != "" {
		hospital.HospitalName = req.HospitalName
	}
	if req.Address != "" {
		hospital.Address = req.Address
	}
	if req.City != "" {
		hospital.City = req.City
	}
	if req.Phone != "" {
		hospital.Phone = req.Phone
	}
	if req.ContactPersonName != "" {
		hospital.ContactPersonName = req.ContactPersonName
	}
	if req.ContactPersonPhone != "" {
		hospital.ContactPersonPhone = req.ContactPersonPhone
	}
	if req.Status != "" {
		hospital.Status = req.Status
	}

	if err := u.repo.Update(ctx, hospital); err != nil {
		return nil, errors.New("failed to update hospital")
	}

	return hospital, nil
}
