package Usecase

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"errors"
	"time"

	"github.com/google/uuid"
)

type DonorBloodRequestUsecase struct {
	repo Interfaces.IDonorBloodRequestRepository
}

func NewDonorBloodRequestUsecase(r Interfaces.IDonorBloodRequestRepository) *DonorBloodRequestUsecase {
	return &DonorBloodRequestUsecase{repo: r}
}

////////////////////////
// CREATE REQUEST
////////////////////////

func (u *DonorBloodRequestUsecase) CreateRequest(
	userID string,
	quantity int,
	reason string,
	hospitalName string,
	hospitalAddress string,
	hospitalPhone string,
) error {

	donorID, err := u.repo.GetDonorIDByUserID(userID)
	if err != nil {
		return err
	}

	// Gate: donor must have at least one successful donation (blood cleared into inventory)
	hasSuccessful, err := u.repo.HasSuccessfulDonation(donorID)
	if err != nil {
		return err
	}
	if !hasSuccessful {
		return errors.New("at least perform one successful donation to request blood")
	}

	donor, err := u.repo.GetDonorProfile(donorID)
	if err != nil {
		return err
	}

	req := &Domain.DonorBloodRequest{
		RequestID: uuid.New().String(),
		DonorID:   donorID,

		DonorName:    donor.FullName,
		DonorEmail:   donor.Email,
		DonorPhone:   donor.Phone,
		DonorAddress: donor.Address,
		BloodType:    donor.BloodType,

		QuantityML: quantity,
		Reason:     reason,

		HospitalName:    hospitalName,
		HospitalAddress: hospitalAddress,
		HospitalPhone:   hospitalPhone,

		Status:    "PENDING",
		CreatedAt: time.Now(),
	}

	return u.repo.Create(req)
}

////////////////////////
// APPROVE REQUEST
//
// Returns (message, error).
// Three scenarios based on how many ML are available in inventory:
//   - reservedML == 0                  → auto-REJECTED,           message: "no enough blood in the inventory"
//   - 0 < reservedML < req.QuantityML  → PARTIALLY APPROVED,      message: "partially approved"
//   - reservedML >= req.QuantityML     → APPROVED,                 message: "fully approved"
//
// Reservation is scoped by the unique requestID, so each donor's reserved
// blood units are completely isolated from other donors' reservations.
////////////////////////

func (u *DonorBloodRequestUsecase) ApproveRequest(requestID string) (*Domain.DonorBloodRequest, string, error) {

	req, err := u.repo.GetByID(requestID)
	if err != nil {
		return nil, "", err
	}

	// Prevent double processing
	switch req.Status {
	case "APPROVED", "PARTIALLY APPROVED", "FULFILLED", "PARTIALLY FULFILLED":
		return nil, "", errors.New("request already processed")
	case "REJECTED":
		return nil, "", errors.New("request already rejected")
	}

	// ReserveBloodUnits returns how many ML were actually reserved.
	// It only touches AVAILABLE (non-expired) units and stamps each unit
	// with donor_request_id = requestID — so Donor1 and Donor2 never share units.
	reservedML, err := u.repo.ReserveBloodUnits(requestID, req.BloodType, req.QuantityML)
	if err != nil {
		return nil, "", err
	}

	var newStatus, message string

	// Scenario 1: no blood available at all
	if reservedML == 0 {
		newStatus = "REJECTED"
		message = "no enough blood in the inventory"
	} else if reservedML < req.QuantityML {
		// Scenario 2: partial blood available
		newStatus = "PARTIALLY APPROVED"
		message = "partially approved"
	} else {
		// Scenario 3: fully available
		newStatus = "APPROVED"
		message = "fully approved"
	}

	if err := u.repo.UpdateStatus(requestID, newStatus); err != nil {
		return nil, "", err
	}

	// Re-fetch updated request to return latest state
	updated, err := u.repo.GetByID(requestID)
	if err != nil {
		return nil, "", err
	}
	return updated, message, nil
}

////////////////////////
// REJECT REQUEST
////////////////////////

func (u *DonorBloodRequestUsecase) RejectRequest(requestID string) (*Domain.DonorBloodRequest, error) {

	req, err := u.repo.GetByID(requestID)
	if err != nil {
		return nil, err
	}

	switch req.Status {
	case "FULFILLED", "PARTIALLY FULFILLED":
		return nil, errors.New("cannot reject a fulfilled request")
	case "REJECTED":
		return nil, errors.New("request already rejected")
	}

	if err := u.repo.UpdateStatus(requestID, "REJECTED"); err != nil {
		return nil, err
	}

	updated, err := u.repo.GetByID(requestID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

////////////////////////
// FULFILL REQUEST
//
// Admin clicks "Fulfilled" after blood physically leaves the blood bank.
// Only APPROVED or PARTIALLY APPROVED requests can be fulfilled.
// Reserved units for this request (matched by donor_request_id = requestID)
// are marked USED — other donors' reservations are never touched.
//
// Status transitions:
//   APPROVED          → FULFILLED
//   PARTIALLY APPROVED → PARTIALLY FULFILLED
////////////////////////

func (u *DonorBloodRequestUsecase) FulfillRequest(requestID string) (*Domain.DonorBloodRequest, error) {

	req, err := u.repo.GetByID(requestID)
	if err != nil {
		return nil, err
	}

	switch req.Status {
	case "REJECTED":
		return nil, errors.New("cannot fulfill a rejected request")
	case "FULFILLED", "PARTIALLY FULFILLED":
		return nil, errors.New("request already fulfilled")
	case "PENDING":
		return nil, errors.New("request has not been approved yet")
	}

	// Mark all blood units reserved for THIS specific request as USED
	if err := u.repo.MarkReservedAsUsed(requestID); err != nil {
		return nil, err
	}

	nextStatus := "FULFILLED"
	if req.Status == "PARTIALLY APPROVED" {
		nextStatus = "PARTIALLY FULFILLED"
	}

	if err := u.repo.UpdateStatus(requestID, nextStatus); err != nil {
		return nil, err
	}

	updated, err := u.repo.GetByID(requestID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

////////////////////////
// GET METHODS
////////////////////////

func (u *DonorBloodRequestUsecase) GetAllRequests() ([]Domain.DonorBloodRequest, error) {
	return u.repo.GetAll()
}

// GetAllAdminRequests returns filtered donor requests sorted by successful
// donation count descending — donors who donated more appear first (higher priority).
func (u *DonorBloodRequestUsecase) GetAllAdminRequests(filter Domain.DonorBloodRequestFilter) ([]Domain.DonorBloodRequest, error) {
	return u.repo.GetAllAdmin(filter)
}

func (u *DonorBloodRequestUsecase) GetMyRequests(userID string, filter Domain.DonorBloodRequestFilter) ([]Domain.DonorBloodRequest, error) {
	donorID, err := u.repo.GetDonorIDByUserID(userID)
	if err != nil {
		return nil, err
	}
	return u.repo.GetByDonorID(donorID, filter)
}

////////////////////////
// EXPIRE STALE RESERVATIONS
//
// Any RESERVED blood unit held for > 24 hours is released back to AVAILABLE,
// and its linked donor request is auto-REJECTED.
// Should be called by a periodic background job (e.g., every 5 minutes).
////////////////////////

func (u *DonorBloodRequestUsecase) ExpireStaleRequests() error {
	return u.repo.ExpireStaleReservations()
}