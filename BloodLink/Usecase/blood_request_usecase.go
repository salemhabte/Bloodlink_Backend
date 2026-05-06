package Usecase

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"bloodlink/Infrastructure"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type bloodRequestUsecase struct {
	repo          Interfaces.IBloodRequestRepository
	hospitalRepo  Interfaces.IHospitalRepository
	inventoryRepo Interfaces.IBloodInventoryRepository
	emergencyUC   Interfaces.IEmergencyRequestUsecase
}

func NewBloodRequestUsecase(
	repo Interfaces.IBloodRequestRepository,
	hospitalRepo Interfaces.IHospitalRepository,
	inventoryRepo Interfaces.IBloodInventoryRepository,
	emergencyUC Interfaces.IEmergencyRequestUsecase,
) Interfaces.IBloodRequestUsecase {
	return &bloodRequestUsecase{
		repo:          repo,
		hospitalRepo:  hospitalRepo,
		inventoryRepo: inventoryRepo,
		emergencyUC:   emergencyUC,
	}
}

func (u *bloodRequestUsecase) CreateBloodRequest(req *Domain.CreateBloodRequestDTO, hospitalAdminUserID string) error {
	hospital_id, err := u.getHospitalIDForAdmin(hospitalAdminUserID)
	if err != nil {
		return err
	}

	contracts, err := u.hospitalRepo.GetContractsByHospitalID(hospital_id)
	if err != nil {
		return err
	}

	hasFinalizedContract := false
	for _, c := range contracts {
		if c.Status == "FINALIZED" {
			hasFinalizedContract = true
			break
		}
	}

	if !hasFinalizedContract {
		return errors.New("cannot create blood request: hospital does not have a finalized contract")
	}

	requestID := uuid.New().String()
	br := &Domain.BloodRequest{
		RequestID:    requestID,
		HospitalID:   hospital_id,
		BloodType:    req.BloodType,
		Quantity:     req.Quantity,
		UrgencyLevel: req.UrgencyLevel,
		Status:       Domain.BloodRequestStatusPending,
		CreatedAt:    time.Now(),
	}

	err = u.repo.CreateRequest(br)
	if err != nil {
		return err
	}

	hospital, err := u.hospitalRepo.GetHospitalByID(hospital_id)
	hospitalName := "A hospital"
	hospitalLocation := "Unknown"
	if err == nil {
		hospitalName = hospital.Name
		hospitalLocation = hospital.Address
	}

	available, err := u.inventoryRepo.CountAvailableUnitsByBloodType(req.BloodType)
	if err == nil {
		if available < req.Quantity {
			_ = u.emergencyUC.TriggerEmergency(requestID, req.BloodType, req.Quantity, req.UrgencyLevel, hospitalName, hospitalLocation)
		}
	}

	adminEmail := "admin@bloodlink.com"
	go func() {
		subject := fmt.Sprintf("New %s Blood Request from %s", req.UrgencyLevel, hospitalName)
		content := fmt.Sprintf("Hospital <b>%s</b> has requested %d units of %s blood.<br><br>Urgency: <b>%s</b>.<br>Location: <b>%s</b>.<br>Please review this request on the admin dashboard.", hospitalName, req.Quantity, req.BloodType, req.UrgencyLevel, hospitalLocation)
		_ = Infrastructure.SendBloodRequestNotification(adminEmail, subject, content)
	}()

	return nil
}

func (u *bloodRequestUsecase) getHospitalIDForAdmin(userID string) (string, error) {
	admin, err := u.hospitalRepo.GetHospitalAdminByUserID(userID)
	if err != nil {
		return "", errors.New("hospital administrator details not found")
	}
	return admin.HospitalID, nil
}

func (u *bloodRequestUsecase) GetHospitalRequests(filter Domain.BloodRequestFilter) ([]Domain.BloodRequestResponse, error) {
	hospital_id, err := u.getHospitalIDForAdmin(filter.HospitalID)
	if err != nil {
		return nil, err
	}
	filter.HospitalID = hospital_id
	return u.repo.GetRequestsByHospital(filter)
}

func (u *bloodRequestUsecase) GetAllRequests() ([]Domain.BloodRequestResponse, error) {
	return u.repo.GetAllRequests()
}

func (u *bloodRequestUsecase) ApproveRequest(requestID string) (*Domain.ApproveRequestResult, error) {
	br, err := u.repo.GetRequestByID(requestID)
	if err != nil {
		return nil, err
	}

	if br.Status != Domain.BloodRequestStatusPending {
		return nil, fmt.Errorf("cannot approve: request is already %s", br.Status)
	}

	// 1. Reserve units (FIFO by expiry is handled in Repo)
	reservedUnits, err := u.inventoryRepo.ReserveUnitsForHospital(br.BloodType, br.Quantity, br.HospitalID, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve blood: %v", err)
	}

	totalReservedVolume := 0
	fulfilledCount := len(reservedUnits)
	var reservedInfo []Domain.ReservedUnitInfo

	for _, unit := range reservedUnits {
		totalReservedVolume += unit.VolumeML
		reservedInfo = append(reservedInfo, Domain.ReservedUnitInfo{
			BloodUnitID:    unit.BloodUnitID,
			BloodType:      unit.BloodType,
			VolumeML:       unit.VolumeML,
			ExpirationDate: unit.ExpirationDate,
		})
	}

	var status string
	var message string
	var notes string

	if fulfilledCount == 0 {
		// a. No blood in inventory
		status = Domain.BloodRequestStatusRejected
		message = "No blood"
		notes = "Automatically rejected: No matching blood units available in inventory."
	} else if fulfilledCount >= br.Quantity {
		// b. Fully fulfilled
		status = Domain.BloodRequestStatusFulfilled
		message = "Request fully fulfilled and blood units reserved."
		notes = fmt.Sprintf("Fulfilled %d units. Total Volume: %dML.", fulfilledCount, totalReservedVolume)
	} else {
		// c. Partially fulfilled
		status = Domain.BloodRequestStatusPartiallyFulfilled
		message = fmt.Sprintf("Request partially fulfilled. Reserved %d units.", fulfilledCount)
		
		unitDetails := ""
		for i, ui := range reservedInfo {
			unitDetails += fmt.Sprintf("%dML", ui.VolumeML)
			if i < len(reservedInfo)-1 {
				unitDetails += " + "
			}
		}
		notes = fmt.Sprintf("Partial fulfillment: %s = %dML total.", unitDetails, totalReservedVolume)
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	err = u.repo.UpdateRequestStatusWithDetails(requestID, status, &now, notes, fulfilledCount, totalReservedVolume)
	if err != nil {
		return nil, err
	}

	// Notify Hospital
	go func() {
		hospitalAdminEmail := "hospitaladmin@bloodlink.com"
		subject := fmt.Sprintf("Update on your Blood Request (%s)", br.BloodType)
		content := fmt.Sprintf("Your request for %d units of %s blood has been updated to: <b>%s</b>.<br>Notes: %s", br.Quantity, br.BloodType, status, notes)
		_ = Infrastructure.SendBloodRequestNotification(hospitalAdminEmail, subject, content)
	}()

	return &Domain.ApproveRequestResult{
		Status:         status,
		Message:        message,
		ReservedUnits:  reservedInfo,
		TotalVolumeML:  totalReservedVolume,
		RequestedCount: br.Quantity,
		FulfilledCount: fulfilledCount,
	}, nil
}

func (u *bloodRequestUsecase) RejectRequest(requestID string) error {
	br, err := u.repo.GetRequestByID(requestID)
	if err != nil {
		return err
	}

	if br.Status != Domain.BloodRequestStatusPending {
		return fmt.Errorf("cannot reject: request is already %s", br.Status)
	}

	err = u.repo.UpdateRequestStatus(requestID, Domain.BloodRequestStatusRejected, nil)
	if err != nil {
		return err
	}

	// Notify Hospital
	go func() {
		hospitalAdminEmail := "hospitaladmin@bloodlink.com"
		subject := fmt.Sprintf("Blood Request Rejected (%s)", br.BloodType)
		content := fmt.Sprintf("Your request for %d units of %s blood has been rejected by the Blood Bank Admin.", br.Quantity, br.BloodType)
		_ = Infrastructure.SendBloodRequestNotification(hospitalAdminEmail, subject, content)
	}()

	return nil
}
