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
	repo         Interfaces.IBloodRequestRepository
	hospitalRepo Interfaces.IHospitalRepository
}

func NewBloodRequestUsecase(repo Interfaces.IBloodRequestRepository, hospitalRepo Interfaces.IHospitalRepository) Interfaces.IBloodRequestUsecase {
	return &bloodRequestUsecase{repo: repo, hospitalRepo: hospitalRepo}
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

	// Notify blood bank admins
	hospital, err := u.hospitalRepo.GetHospitalByID(hospital_id)
	hospitalName := "A hospital"
	if err == nil {
		hospitalName = hospital.Name
	}

	// Send Notification to Blood Bank Admin (Assuming static admin email for now)
	adminEmail := "admin@bloodlink.com"
	go func() {
		subject := fmt.Sprintf("New %s Blood Request from %s", req.UrgencyLevel, hospitalName)
		content := fmt.Sprintf("Hospital <b>%s</b> has requested %d units of %s blood.<br><br>Urgency: <b>%s</b>.<br>Please review this request on the admin dashboard.", hospitalName, req.Quantity, req.BloodType, req.UrgencyLevel)
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

func (u *bloodRequestUsecase) GetHospitalRequests(hospitalAdminID string) ([]Domain.BloodRequestResponse, error) {
	hospital_id, err := u.getHospitalIDForAdmin(hospitalAdminID)
	if err != nil {
		return nil, err
	}
	return u.repo.GetRequestsByHospital(hospital_id)
}

func (u *bloodRequestUsecase) GetAllRequests() ([]Domain.BloodRequestResponse, error) {
	return u.repo.GetAllRequests()
}

func (u *bloodRequestUsecase) UpdateStatus(requestID string, req *Domain.UpdateBloodRequestStatusDTO) error {
	br, err := u.repo.GetRequestByID(requestID)
	if err != nil {
		return err
	}

	var approvedAtStr *string
	// If it transitions to APPROVED_PARTIALLY_FULFILLED or FULFILLED
	if req.Status == Domain.BloodRequestStatusPartiallyFulfilled || req.Status == Domain.BloodRequestStatusFulfilled {
		now := time.Now().Format("2006-01-02 15:04:05")
		approvedAtStr = &now
	}

	err = u.repo.UpdateRequestStatus(requestID, req.Status, approvedAtStr)
	if err != nil {
		return err
	}

	// Notify Hospital that status changed
	// We need the admin's email for the hospital.
	hospital, err := u.hospitalRepo.GetHospitalByID(br.HospitalID)
	if err == nil {
		// Ideally we fetch all hospital admins for this hospital, but since we just have one mostly, we just notify an email list.
		// For simplicity we will notify the admin email.
		// Since we lack a direct `GetUsersByHospitalID`, we'll assume a dummy or config email here.
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
