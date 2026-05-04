package Usecase

import (
	"bloodlink/Domain"
	"bloodlink/Repository"
	"errors"
	"time"

	"github.com/google/uuid"
)

type DonorBloodRequestUsecase struct {
	repo *Repository.DonorBloodRequestRepository
}

func NewDonorBloodRequestUsecase(r *Repository.DonorBloodRequestRepository) *DonorBloodRequestUsecase {
	return &DonorBloodRequestUsecase{repo: r}
}

///////////////////////
// CREATE REQUEST
///////////////////////

func (u *DonorBloodRequestUsecase) CreateRequest(
	donorID string,
	quantity int,
	reason string,
) error {

	// 1. check donor eligibility
	count, err := u.repo.CountDonationsByDonor(donorID)
	if err != nil {
		return err
	}
	if count < 1 {
		return errors.New("you must donate at least once before requesting blood")
	}

	// 2. get blood type from system
	bloodType, err := u.repo.GetDonorBloodType(donorID)
	if err != nil {
		return err
	}

	// 3. priority logic
	priority := count * 10

	// 4. create request
	req := &Domain.DonorBloodRequest{
		RequestID:     uuid.New().String(),
		DonorID:       donorID,
		BloodType:     bloodType,
		QuantityML:    quantity,
		Reason:        reason,
		PriorityScore: priority,
		Status:        "PENDING",
		CreatedAt:     time.Now(),
	}

	return u.repo.Create(req)
}

///////////////////////
// APPROVE REQUEST
///////////////////////

func (u *DonorBloodRequestUsecase) ApproveRequest(requestID string) error {

	req, err := u.repo.GetByID(requestID)
	if err != nil {
		return err
	}

	// Status Locking
	if req.Status == "FULFILLED" || req.Status == "REJECTED" {
		return errors.New("cannot approve: request is already " + req.Status)
	}

	// check available blood
	available, err := u.repo.GetAvailableBloodVolume(req.BloodType)
	if err != nil {
		return err
	}

	if available < req.QuantityML {
		// ❌ auto reject
		return u.repo.UpdateStatus(requestID, "REJECTED")
	}

	// ✅ reserve blood units
	err = u.repo.ReserveBlood(req.BloodType, req.QuantityML)
	if err != nil {
		return err
	}

	// update request
	return u.repo.UpdateStatus(requestID, "APPROVED")
}


///////////////////////
// GET ALL REQUESTS
///////////////////////

func (u *DonorBloodRequestUsecase) GetAllRequests() ([]Domain.DonorBloodRequest, error) {
	return u.repo.GetAll()
}
func (u *DonorBloodRequestUsecase) RejectRequest(requestID string) error {
	req, err := u.repo.GetByID(requestID)
	if err != nil {
		return err
	}

	// Status Locking
	if req.Status == "FULFILLED" || req.Status == "REJECTED" {
		return errors.New("cannot reject: request is already " + req.Status)
	}

	return u.repo.UpdateStatus(requestID, "REJECTED")
}
func (u *DonorBloodRequestUsecase) GetMyRequests(donorID string) ([]Domain.DonorBloodRequest, error) {
	return u.repo.GetByDonorID(donorID)
}
func (u *DonorBloodRequestUsecase) FulfillRequest(requestID string) error {

	req, err := u.repo.GetByID(requestID)
	if err != nil {
		return err
	}

	//  prevent invalid flow
	if req.Status == "FULFILLED" || req.Status == "REJECTED" {
		return errors.New("cannot fulfill: request is already " + req.Status)
	}
	if req.Status != "APPROVED" {
		return errors.New("only approved requests can be fulfilled")
	}

	// mark RESERVED → USED
	err = u.repo.MarkReservedAsUsed(req.BloodType, req.QuantityML)
	if err != nil {
		return err
	}

	return u.repo.UpdateStatus(requestID, "FULFILLED")
}