package Usecase

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"bloodlink/Infrastructure"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type emergencyRequestUsecase struct {
	repo          Interfaces.IEmergencyRequestRepository
	inventoryRepo Interfaces.IBloodInventoryRepository
	hospitalRepo  Interfaces.IHospitalRepository
	requestRepo   Interfaces.IBloodRequestRepository
	userRepo      IUserRepository
	profileRepo   IProfileRepository
	notifUC       Interfaces.INotificationUsecase
}

func NewEmergencyRequestUsecase(
	repo Interfaces.IEmergencyRequestRepository,
	inventoryRepo Interfaces.IBloodInventoryRepository,
	hospitalRepo Interfaces.IHospitalRepository,
	requestRepo Interfaces.IBloodRequestRepository,
	userRepo IUserRepository,
	profileRepo IProfileRepository,
	notifUC Interfaces.INotificationUsecase,
) Interfaces.IEmergencyRequestUsecase {
	return &emergencyRequestUsecase{
		repo:          repo,
		inventoryRepo: inventoryRepo,
		hospitalRepo:  hospitalRepo,
		requestRepo:   requestRepo,
		userRepo:      userRepo,
		profileRepo:   profileRepo,
		notifUC:       notifUC,
	}
}

func (u *emergencyRequestUsecase) TriggerEmergency(requestID string, bloodType string, quantity int, urgencyLevel string, hospitalName string, location string, latitude float64, longitude float64) error {
	// Check if an emergency for this request already exists
	existing, err := u.repo.GetByRequestID(requestID)
	if err == nil && existing != nil {
		return nil // Already triggered
	}

	emergency := &Domain.EmergencyRequest{
		EmergencyID:       uuid.New().String(),
		RequestID:         &requestID,
		BloodType:         bloodType,
		QuantityRequired:  quantity,
		QuantityFulfilled: 0,
		UrgencyLevel:      urgencyLevel,
		HospitalName:      hospitalName,
		Location:          location,
		Latitude:          latitude,
		Longitude:         longitude,
		Status:            Domain.EmergencyStatusPending,
		IsManual:          false,
		CreatedAt:         time.Now(),
		EndDate:           nil, // Will be set upon publishing if not provided
	}

	err = u.repo.Create(emergency)
	if err != nil {
		return err
	}

	// Notify Blood Bank Admin
	adminEmail := "admin@bloodlink.com"
	go func() {
		subject := fmt.Sprintf("URGENT: %s Emergency Triggered - %s", urgencyLevel, hospitalName)
		content := fmt.Sprintf("A hospital request for %d units of %s blood at <b>%s</b> (%s) has triggered an emergency.<br>Urgency: <b>%s</b>.<br>Please review and publish this emergency request.", quantity, bloodType, hospitalName, location, urgencyLevel)
		_ = Infrastructure.SendBloodRequestNotification(adminEmail, subject, content)
	}()

	return nil
}

func (u *emergencyRequestUsecase) PublishEmergency(id string) error {
	req, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}

	if req.Status != Domain.EmergencyStatusPending {
		return errors.New("only pending emergencies can be published")
	}

	err = u.repo.UpdateStatus(id, Domain.EmergencyStatusPublished)
	if err != nil {
		return err
	}

	// Trigger notifications to donors of this blood type in the hospital's area
	go func() {
		ctx := context.Background()
		hospitalAddress := req.Location
		hospitalName := req.HospitalName

		if hospitalName == "" {
			hospitalName = "a nearby hospital"
		}

		if req.Latitude != 0 && req.Longitude != 0 {
			// Get donors nearby (e.g., 20km radius)
			radiusKm := 20.0
			donors, err := u.userRepo.GetDonorsNearby(ctx, req.BloodType, req.Latitude, req.Longitude, radiusKm)
			if err == nil {
				for _, donor := range donors {
					subject := fmt.Sprintf("URGENT: %s Blood Emergency nearby", req.BloodType)
					content := fmt.Sprintf("Hello %s,<br><br><b>%s</b> urgently needs <b>%d units</b> of <b>%s</b> blood.<br>Urgency: <b>%s</b>.<br>Since you are within %.1f km, your donation could save a life!<br><br>Please visit the hospital at <b>%s</b> or contact us for more details.", donor.FullName, hospitalName, req.QuantityRequired, req.BloodType, req.UrgencyLevel, radiusKm, hospitalAddress)
					_ = Infrastructure.SendBloodRequestNotification(donor.Email, subject, content)
					_ = u.notifUC.SendNotification(donor.UserID, "EMERGENCY", "URGENT: Blood Emergency", fmt.Sprintf("%s needs %s blood.", hospitalName, req.BloodType))
				}
			}
		}
	}()

	return nil
}

func (u *emergencyRequestUsecase) RejectEmergency(id string) error {
	req, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}

	if req.Status != Domain.EmergencyStatusPending {
		return errors.New("only pending emergencies can be rejected")
	}

	return u.repo.UpdateStatus(id, Domain.EmergencyStatusRejected)
}

func (u *emergencyRequestUsecase) CreateManualEmergency(dtos []Domain.CreateEmergencyRequestDTO) error {
	for _, dto := range dtos {
		emergency := &Domain.EmergencyRequest{
			EmergencyID:       uuid.New().String(),
			BloodType:         dto.BloodType,
			QuantityRequired:  dto.QuantityRequired,
			QuantityFulfilled: 0,
			UrgencyLevel:      "emergency",
			HospitalName:      dto.HospitalName,
			Location:          dto.Location,
			Latitude:          dto.Latitude,
			Longitude:         dto.Longitude,
			Status:            Domain.EmergencyStatusPublished, // Manual ones are published immediately
			IsManual:          true,
			CreatedAt:         time.Now(),
		}

		if dto.EndDate != "" {
			ed, err := time.Parse("2006-01-02", dto.EndDate)
			if err != nil {
				return errors.New("invalid end_date format, must be YYYY-MM-DD")
			}
			today := time.Now().Truncate(24 * time.Hour)
			if ed.Before(today) {
				return errors.New("end_date cannot be in the past")
			}
			emergency.EndDate = &ed
		}

		now := time.Now()
		emergency.PublishedAt = &now

		err := u.repo.Create(emergency)
		if err != nil {
			return err
		}

		// Notify nearby donors if location is provided
		if emergency.Latitude != 0 && emergency.Longitude != 0 {
			go func(e *Domain.EmergencyRequest) {
				ctx := context.Background()
				radiusKm := 20.0
				donors, err := u.userRepo.GetDonorsNearby(ctx, e.BloodType, e.Latitude, e.Longitude, radiusKm)
				if err == nil {
					for _, donor := range donors {
						subject := fmt.Sprintf("URGENT: %s Blood Emergency nearby", e.BloodType)
						content := fmt.Sprintf("Hello %s,<br><br><b>%s</b> urgently needs <b>%d units</b> of <b>%s</b> blood.<br>Urgency: <b>%s</b>.<br>Since you are within %.1f km, your donation could save a life!<br><br>Please visit the hospital at <b>%s</b> or contact us for more details.", donor.FullName, e.HospitalName, e.QuantityRequired, e.BloodType, e.UrgencyLevel, radiusKm, e.Location)
						_ = Infrastructure.SendBloodRequestNotification(donor.Email, subject, content)
						_ = u.notifUC.SendNotification(donor.UserID, "EMERGENCY", "URGENT: Blood Emergency", fmt.Sprintf("%s needs %s blood.", e.HospitalName, e.BloodType))
					}
				}
			}(emergency)
		}
	}
	return nil
}


func (u *emergencyRequestUsecase) GetAllEmergencies(filter Domain.EmergencyRequestFilter) (*Domain.EmergencyListResponse, error) {
	emergencies, err := u.repo.GetAll(filter)
	if err != nil {
		return nil, err
	}

	var analytics Domain.EmergencyAnalytics
	analytics.TotalRequests = len(emergencies)
	for _, e := range emergencies {
		switch e.Status {
		case Domain.EmergencyStatusPublished:
			analytics.TotalPublished++
			analytics.TotalActive++
		case Domain.EmergencyStatusCompleted:
			analytics.TotalEnded++
		}
	}

	return &Domain.EmergencyListResponse{
		Emergencies: emergencies,
		Analytics:   analytics,
	}, nil
}

func (u *emergencyRequestUsecase) GetPublishedEmergencies() ([]Domain.EmergencyRequest, error) {
	return u.repo.GetActive()
}

func (u *emergencyRequestUsecase) GetEmergenciesForDonor(userID string) ([]Domain.EmergencyRequest, error) {
	ctx := context.Background()

	// Get profile for location
	profile, err := u.profileRepo.GetProfileByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil || profile.Latitude == nil || profile.Longitude == nil {
		return nil, errors.New("donor profile or location not found")
	}

	// Get donor's confirmed blood type
	bloodType := ""
	donor, err := u.userRepo.GetDonorByUserID(ctx, userID)
	if err == nil && donor != nil {
		bloodType = donor.BloodType // if empty string, GetNearby returns all blood types
	}

	// Find emergencies within 20km, filtered by blood type if known
	return u.repo.GetNearby(*profile.Latitude, *profile.Longitude, 20.0, bloodType)
}

func (u *emergencyRequestUsecase) MarkCompletedEmergencies() error {
	return u.repo.MarkCompletedEmergencies()
}
