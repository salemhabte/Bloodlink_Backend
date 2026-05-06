package Domain

import "bloodlink/Domain"

type INotificationRepository interface {
	CreateNotification(notification *Domain.Notification) error
	GetNotificationsByUserID(userID string) ([]Domain.Notification, error)
	MarkAsRead(notificationID string, userID string) error
	MarkAllAsRead(userID string) error

	// Helper methods to get user IDs by role
	GetUserIDsByRole(role string) ([]string, error)
	GetUserIDByHospitalID(hospitalID string) (string, error)
	GetUserIDByDonorID(donorID string) (string, error)
}

type INotificationUsecase interface {
	SendNotification(userID, notifType, title, message string) error
	SendToRole(role, notifType, title, message string) error
	SendToHospital(hospitalID, notifType, title, message string) error
	SendToDonor(donorID, notifType, title, message string) error

	GetMyNotifications(userID string) ([]Domain.Notification, error)
	MarkAsRead(notificationID string, userID string) error
	MarkAllAsRead(userID string) error
}
