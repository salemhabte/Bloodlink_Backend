package Usecase

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
)

type AuditLogUsecase struct {
	repo Interfaces.IAdminAuditLogRepository
}

func NewAuditLogUsecase(repo Interfaces.IAdminAuditLogRepository) *AuditLogUsecase {
	return &AuditLogUsecase{repo: repo}
}

// LogAction is a helper to quickly insert a log entry
func (u *AuditLogUsecase) LogAction(adminID, action, targetType, targetID, details string) error {
	log := &Domain.AdminAuditLog{
		AdminID:    adminID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Details:    details,
	}
	return u.repo.CreateLog(log)
}

func (u *AuditLogUsecase) GetLogs(filter Domain.AuditLogFilter) (*Domain.AuditLogListResponse, error) {
	logs, total, err := u.repo.GetLogs(filter)
	if err != nil {
		return nil, err
	}

	return &Domain.AuditLogListResponse{
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
		Logs:  logs,
	}, nil
}

func (u *AuditLogUsecase) GetLogByID(id string) (*Domain.AdminAuditLog, error) {
	return u.repo.GetLogByID(id)
}

func (u *AuditLogUsecase) DeleteLog(id string) error {
	return u.repo.DeleteLog(id)
}

func (u *AuditLogUsecase) ExportLogs(filter Domain.AuditLogFilter) ([]Domain.AdminAuditLog, error) {
	// For export, we might want to get all matching logs regardless of pagination,
	// but to be safe we can use a high limit (e.g., 10000).
	filter.Page = 1
	filter.Limit = 10000
	logs, _, err := u.repo.GetLogs(filter)
	return logs, err
}
