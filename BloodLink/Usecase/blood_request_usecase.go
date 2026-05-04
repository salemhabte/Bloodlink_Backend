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
	// Need to find which Hospital this user belongs to
	// We lack a direct GetHospitalByAdminUserID repo method, but we can assume we'll either make one or we fake it.
	// Actually, wait, the auth claims might inject hospital_id, but if it only has user_id...
	// Let's assume there's a way to find hospital via admin. For now, since HospitalAdmin has `user_id` and `hospital_id`,
	// we would query `hospital_admins` table. Let's do a direct look up if possible.
	hospital_id, err := u.getHospitalIDForAdmin(hospitalAdminUserID)
	if err != nil {
		return err
	}

	// CHECK FOR FINALIZED CONTRACT
	contracts, err := u.hospitalRepo.GetContractsByHospitalID(hospital_id)
	if err != nil {
		return err
	}

	hasFinalizedContract := false
	for _, c := range contracts {
		if c.Status == Domain.ContractStatusFinalized {
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

	// Get hospital info for emergency and notification
	hospital, err := u.hospitalRepo.GetHospitalByID(hospital_id)
	hospitalName := "A hospital"
	hospitalLocation := "Unknown"
	if err == nil {
		hospitalName = hospital.Name
		hospitalLocation = hospital.Address
	}

	// CHECK INVENTORY FOR EMERGENCY TRIGGER
	available, err := u.inventoryRepo.CountAvailableUnitsByBloodType(req.BloodType)
	if err == nil {
		if available < req.Quantity {
			// Trigger emergency
			_ = u.emergencyUC.TriggerEmergency(requestID, req.BloodType, req.Quantity, req.UrgencyLevel, hospitalName, hospitalLocation)
		}
	}

	// Send Notification to Blood Bank Admin (Assuming static admin email for now)
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

func (u *bloodRequestUsecase) UpdateStatus(requestID string, req *Domain.UpdateBloodRequestStatusDTO) error {
	br, err := u.repo.GetRequestByID(requestID)
	if err != nil {
		return err
	}

	// Status Locking: Once changed from PENDING, it shouldn't change again
	if br.Status != Domain.BloodRequestStatusPending {
		return fmt.Errorf("cannot update status: request is already %s", br.Status)
	}

	var approvedAtStr *string
	// If it transitions to FULFILLED
	if req.Status == Domain.BloodRequestStatusFulfilled {
		// 1. Check Inventory
		available, err := u.inventoryRepo.CountAvailableUnitsByBloodType(br.BloodType)
		if err != nil {
			return fmt.Errorf("failed to check inventory: %v", err)
		}

		if available < br.Quantity {
			return fmt.Errorf("insufficient inventory: only %d units of %s available, but %d requested", available, br.BloodType, br.Quantity)
		}

		// 2. Decrease Inventory
		err = u.inventoryRepo.ConsumeUnits(br.BloodType, br.Quantity)
		if err != nil {
			return fmt.Errorf("failed to fulfill inventory: %v", err)
		}

		now := time.Now().Format("2006-01-02 15:04:05")
		approvedAtStr = &now
	} else if req.Status == Domain.BloodRequestStatusPartiallyFulfilled {
		available, err := u.inventoryRepo.CountAvailableUnitsByBloodType(br.BloodType)
		if err != nil {
			return fmt.Errorf("failed to check inventory: %v", err)
		}

		if available > 0 {
			toConsume := available
			if toConsume > br.Quantity {
				toConsume = br.Quantity
			}
			err = u.inventoryRepo.ConsumeUnits(br.BloodType, toConsume)
			if err != nil {
				return fmt.Errorf("failed to partially fulfill inventory: %v", err)
			}
		}

		now := time.Now().Format("2006-01-02 15:04:05")
		approvedAtStr = &now
	}

	err = u.repo.UpdateRequestStatus(requestID, req.Status, approvedAtStr)
	if err != nil {
		return err
	}

	// Notify Hospital that status changed
	hospital, err := u.hospitalRepo.GetHospitalByID(br.HospitalID)
	if err == nil {
		hospitalAdminEmail := "hospitaladmin@bloodlink.com"

		go func() {
			subject := fmt.Sprintf("Update on your Blood Request (%s)", br.BloodType)
			content := fmt.Sprintf("Your request for %d units of %s blood has been updated to: <b>%s</b>.", br.Quantity, br.BloodType, req.Status)
			log.Printf("Notifying %s: %s", hospital.Name, subject)
			_ = Infrastructure.SendBloodRequestNotification(hospitalAdminEmail, subject, content)
		}()
	}

	return nil
}
