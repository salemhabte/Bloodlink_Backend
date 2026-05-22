package Domain

import "bloodlink/Domain"

type IAdminAuditLogRepository interface {
	CreateLog(log *Domain.AdminAuditLog) error
	GetLogs(filter Domain.AuditLogFilter) ([]Domain.AdminAuditLog, int, error)
	GetLogByID(id string) (*Domain.AdminAuditLog, error)
	DeleteLog(id string) error
}
