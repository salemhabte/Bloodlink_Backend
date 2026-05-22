package Domain

import "time"

type AdminAuditLog struct {
	LogID      string    `json:"log_id" db:"log_id"`
	AdminID    string    `json:"admin_id" db:"admin_id"`
	Action     string    `json:"action" db:"action"`
	TargetType string    `json:"target_type" db:"target_type"`
	TargetID   string    `json:"target_id" db:"target_id"`
	Details    string    `json:"details" db:"details"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type AuditLogFilter struct {
	Action     string `json:"action"`
	TargetType string `json:"target_type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

type AuditLogListResponse struct {
	Total int             `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
	Logs  []AdminAuditLog `json:"logs"`
}
