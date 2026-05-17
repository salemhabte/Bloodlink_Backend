package Usecase

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"bloodlink/Infrastructure"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type bloodRequestUsecase struct {
	repo          Interfaces.IBloodRequestRepository
	hospitalRepo  Interfaces.IHospitalRepository
	inventoryRepo Interfaces.IBloodInventoryRepository
	emergencyUC   Interfaces.IEmergencyRequestUsecase
	notifUC       Interfaces.INotificationUsecase
}

func NewBloodRequestUsecase(
	repo Interfaces.IBloodRequestRepository,
	hospitalRepo Interfaces.IHospitalRepository,
	inventoryRepo Interfaces.IBloodInventoryRepository,
	emergencyUC Interfaces.IEmergencyRequestUsecase,
	notifUC Interfaces.INotificationUsecase,
) Interfaces.IBloodRequestUsecase {
	return &bloodRequestUsecase{
		repo:          repo,
		hospitalRepo:  hospitalRepo,
		inventoryRepo: inventoryRepo,
		emergencyUC:   emergencyUC,
		notifUC:       notifUC,
	}
}

func (u *bloodRequestUsecase) CreateBloodRequest(req *Domain.CreateBloodRequestBatchDTO, hospitalAdminUserID string) error {
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

	hospital, err := u.hospitalRepo.GetHospitalByID(hospital_id)
	hospitalName := "A hospital"
	hospitalLocation := "Unknown"
	if err == nil {
		hospitalName = hospital.Name
		hospitalLocation = hospital.Address
	}

	for _, item := range req.Requests {
		requestID := uuid.New().String()
		br := &Domain.BloodRequest{
			RequestID:    requestID,
			HospitalID:   hospital_id,
			BloodType:    item.BloodType,
			Component:    item.Component,
			Quantity:     item.Quantity,
			UrgencyLevel: req.UrgencyLevel,
			Status:       Domain.BloodRequestStatusPending,
			CreatedAt:    time.Now(),
		}

		err = u.repo.CreateRequest(br)
		if err != nil {
			return err
		}

		available, err := u.inventoryRepo.CountAvailableUnitsByBloodType(item.BloodType)
		if err == nil {
			if available < item.Quantity && req.UrgencyLevel == "emergency" {
				_ = u.emergencyUC.TriggerEmergency(requestID, item.BloodType, item.Quantity, req.UrgencyLevel, hospitalName, hospitalLocation, hospital.Latitude, hospital.Longitude)
			}
		}

		adminEmail := "admin@bloodlink.com"
		go func(rID string, bType string, qty int, comp string) {
			subject := fmt.Sprintf("New %s Blood Request from %s", req.UrgencyLevel, hospitalName)
			content := fmt.Sprintf("Hospital <b>%s</b> has requested %d units of %s (%s).<br><br>Urgency: <b>%s</b>.<br>Location: <b>%s</b>.<br>Please review this request on the admin dashboard.", hospitalName, qty, bType, comp, req.UrgencyLevel, hospitalLocation)
			_ = Infrastructure.SendBloodRequestNotification(adminEmail, subject, content)
			_ = u.notifUC.SendToRole(Domain.RoleBloodBankAdmin, "BLOOD_REQUEST", "New Blood Request", fmt.Sprintf("%s has requested %d units of %s (%s)", hospitalName, qty, bType, comp))
		}(requestID, item.BloodType, item.Quantity, item.Component)
	}

	return nil
}

func (u *bloodRequestUsecase) getHospitalIDForAdmin(userID string) (string, error) {
	admin, err := u.hospitalRepo.GetHospitalAdminByUserID(userID)
	if err != nil {
		return "", errors.New("hospital administrator details not found")
	}
	return admin.HospitalID, nil
}

func (u *bloodRequestUsecase) GetHospitalRequests(filter Domain.BloodRequestFilter) (*Domain.BloodRequestListResponse, error) {
	hospital_id, err := u.getHospitalIDForAdmin(filter.HospitalID)
	if err != nil {
		return nil, err
	}
	filter.HospitalID = hospital_id
	requests, err := u.repo.GetRequestsByHospital(filter)
	if err != nil {
		return nil, err
	}

	return &Domain.BloodRequestListResponse{
		Requests:  requests,
		Analytics: u.calculateAnalytics(requests),
	}, nil
}

func (u *bloodRequestUsecase) GetAllRequests(filter Domain.BloodRequestFilter) (*Domain.BloodRequestListResponse, error) {
	requests, err := u.repo.GetAllRequests(filter)
	if err != nil {
		return nil, err
	}

	return &Domain.BloodRequestListResponse{
		Requests:  requests,
		Analytics: u.calculateAnalytics(requests),
	}, nil
}

func (u *bloodRequestUsecase) calculateAnalytics(requests []Domain.BloodRequestResponse) Domain.SummaryAnalytics {
	var analytics Domain.SummaryAnalytics
	analytics.TotalRequests = len(requests)
	for _, r := range requests {
		switch r.Status {
		case Domain.BloodRequestStatusFulfilled, Domain.BloodRequestStatusPartiallyFulfilled:
			analytics.TotalFulfilled++
		case Domain.BloodRequestStatusPending:
			analytics.TotalPending++
		case Domain.BloodRequestStatusRejected:
			analytics.TotalCancelled++
		}
	}
	return analytics
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
	reservedUnits, err := u.inventoryRepo.ReserveUnitsForHospital(br.BloodType, br.Component, br.Quantity, br.HospitalID, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to reserve blood: %v", err)
	}

	totalReservedQuantity := 0
	fulfilledCount := len(reservedUnits)
	var reservedInfo []Domain.ReservedUnitInfo

	for _, unit := range reservedUnits {
		totalReservedQuantity += unit.QuantityML
		reservedInfo = append(reservedInfo, Domain.ReservedUnitInfo{
			BloodUnitID:    unit.BloodUnitID,
			BloodType:      unit.BloodType,
			QuantityML:       unit.QuantityML,
			ExpirationDate: unit.ExpirationDate,
		})
	}

	var status string
	var message string
	var notes string

	if totalReservedQuantity == 0 {
		// a. No blood in inventory
		status = Domain.BloodRequestStatusRejected
		message = "No blood"
		notes = "Automatically rejected: No available units of the requested blood type in inventory."
	} else if fulfilledCount >= br.Quantity {
		// b. Fully fulfilled
		status = Domain.BloodRequestStatusFulfilled
		message = "Request fully fulfilled and blood units reserved."
		notes = fmt.Sprintf("Fully fulfilled. Reserved %d units totaling %dML.", fulfilledCount, totalReservedQuantity)
	} else {
		// c. Partially fulfilled
		status = Domain.BloodRequestStatusPartiallyFulfilled
		message = fmt.Sprintf("Request partially fulfilled. Reserved %d units totaling %dML.", fulfilledCount, totalReservedQuantity)

		unitDetails := ""
		for i, ui := range reservedInfo {
			unitDetails += fmt.Sprintf("%dML", ui.QuantityML)
			if i < len(reservedInfo)-1 {
				unitDetails += " + "
			}
		}
		notes = fmt.Sprintf("Partial fulfillment: %s = %dML total.", unitDetails, totalReservedQuantity)
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	err = u.repo.UpdateRequestStatusWithDetails(requestID, status, &now, notes, fulfilledCount, totalReservedQuantity)
	if err != nil {
		return nil, err
	}

	// Notify Hospital
	go func() {
		hospitalAdminEmail := "hospitaladmin@bloodlink.com"
		subject := fmt.Sprintf("Update on your Blood Request (%s)", br.BloodType)
		content := fmt.Sprintf("Your request for %d units of %s blood has been updated to: <b>%s</b>.<br>Notes: %s", br.Quantity, br.BloodType, status, notes)
		_ = Infrastructure.SendBloodRequestNotification(hospitalAdminEmail, subject, content)
		_ = u.notifUC.SendToHospital(br.HospitalID, "BLOOD_REQUEST", "Blood Request Update", fmt.Sprintf("Your request for %s blood has been %s", br.BloodType, status))
	}()

	return &Domain.ApproveRequestResult{
		Status:         status,
		Message:        message,
		ReservedUnits:  reservedInfo,
		TotalQuantityML:  totalReservedQuantity,
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
		_ = u.notifUC.SendToHospital(br.HospitalID, "BLOOD_REQUEST", "Blood Request Rejected", fmt.Sprintf("Your request for %s blood was rejected", br.BloodType))
	}()

	return nil
}
