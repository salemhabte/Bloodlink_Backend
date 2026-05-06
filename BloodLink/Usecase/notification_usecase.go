package Usecase

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"log"
	"time"

	"github.com/google/uuid"
)

type notificationUsecase struct {
	repo Interfaces.INotificationRepository
}

func NewNotificationUsecase(repo Interfaces.INotificationRepository) Interfaces.INotificationUsecase {
	return &notificationUsecase{repo: repo}
}

func (u *notificationUsecase) SendNotification(userID, notifType, title, message string) error {
	notif := &Domain.Notification{
		NotificationID: uuid.New().String(),
		UserID:         userID,
		Type:           notifType,
		Title:          title,
		Message:        message,
		IsRead:         false,
		CreatedAt:      time.Now(),
	}
	return u.repo.CreateNotification(notif)
}

func (u *notificationUsecase) SendToRole(role, notifType, title, message string) error {
	userIDs, err := u.repo.GetUserIDsByRole(role)
	if err != nil {
		log.Printf("[NOTIF ERROR] Failed to get user IDs for role %s: %v", role, err)
		return err
	}

	log.Printf("[NOTIF] Sending %s notification to %d users with role %s", notifType, len(userIDs), role)

	for _, id := range userIDs {
		if err := u.SendNotification(id, notifType, title, message); err != nil {
			log.Printf("[NOTIF ERROR] Failed to send notification to user %s: %v", id, err)
		}
	}
	return nil
}

func (u *notificationUsecase) SendToHospital(hospitalID, notifType, title, message string) error {
	userID, err := u.repo.GetUserIDByHospitalID(hospitalID)
	if err != nil {
		return err // No admin to notify
	}
	return u.SendNotification(userID, notifType, title, message)
}

func (u *notificationUsecase) SendToDonor(donorID, notifType, title, message string) error {
	userID, err := u.repo.GetUserIDByDonorID(donorID)
	if err != nil {
		return err
	}
	return u.SendNotification(userID, notifType, title, message)
}

func (u *notificationUsecase) GetMyNotifications(userID string) ([]Domain.Notification, error) {
	return u.repo.GetNotificationsByUserID(userID)
}

func (u *notificationUsecase) MarkAsRead(notificationID string, userID string) error {
	return u.repo.MarkAsRead(notificationID, userID)
}

func (u *notificationUsecase) MarkAllAsRead(userID string) error {
	return u.repo.MarkAllAsRead(userID)
}
