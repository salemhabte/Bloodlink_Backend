package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
	"errors"
)

type notificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) Interfaces.INotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) CreateNotification(notif *Domain.Notification) error {
	query := `
		INSERT INTO notifications (notification_id, user_id, type, title, message, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(query, notif.NotificationID, notif.UserID, notif.Type, notif.Title, notif.Message, notif.IsRead, notif.CreatedAt)
	return err
}

func (r *notificationRepository) GetNotificationsByUserID(userID string) ([]Domain.Notification, error) {
	query := `
		SELECT notification_id, user_id, type, title, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []Domain.Notification
	for rows.Next() {
		var n Domain.Notification
		if err := rows.Scan(&n.NotificationID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifs = append(notifs, n)
	}
	return notifs, nil
}

func (r *notificationRepository) MarkAsRead(notificationID string, userID string) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE notification_id = $1 AND user_id = $2`
	result, err := r.db.Exec(query, notificationID, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("notification not found or unauthorized")
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(userID string) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *notificationRepository) GetUserIDsByRole(role string) ([]string, error) {
	query := `SELECT user_id FROM users WHERE role = $1`
	rows, err := r.db.Query(query, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, nil
}

func (r *notificationRepository) GetUserIDByHospitalID(hospitalID string) (string, error) {
	query := `SELECT user_id FROM hospital_admins WHERE hospital_id = $1 LIMIT 1`
	var userID string
	err := r.db.QueryRow(query, hospitalID).Scan(&userID)
	if err != nil {
		return "", errors.New("hospital admin not found for this hospital")
	}
	return userID, nil
}

func (r *notificationRepository) GetUserIDByDonorID(donorID string) (string, error) {
	query := `SELECT user_id FROM donors WHERE donor_id = $1 LIMIT 1`
	var userID string
	err := r.db.QueryRow(query, donorID).Scan(&userID)
	if err != nil {
		return "", errors.New("user not found for this donor")
	}
	return userID, nil
}
