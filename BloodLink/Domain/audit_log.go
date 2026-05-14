package Domain

import "time"

type AuditLog struct {
	LogID      int       `json:"log_id" db:"log_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	UserName   string    `json:"user_name,omitempty" db:"full_name"`
	Action     string    `json:"action" db:"action"`
	TargetType string    `json:"target_type" db:"target_type"`
	TargetID   string    `json:"target_id" db:"target_id"`
	TargetName string    `json:"target_name" db:"target_name"`
	OldValue   string    `json:"old_value" db:"old_value"`
	NewValue   string    `json:"new_value" db:"new_value"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type AuditLogFilter struct {
	UserID     string `json:"user_id"`
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
}
