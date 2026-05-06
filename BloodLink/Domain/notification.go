package Domain

import "time"

type Notification struct {
	NotificationID string    `json:"notification_id" db:"notification_id"`
	UserID         string    `json:"user_id" db:"user_id"`
	Type           string    `json:"type" db:"type"` // e.g., "CAMPAIGN", "EMERGENCY", "TEST_RESULT", "CONTRACT", "BLOOD_REQUEST", "DONATION"
	Title          string    `json:"title" db:"title"`
	Message        string    `json:"message" db:"message"`
	IsRead         bool      `json:"is_read" db:"is_read"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
