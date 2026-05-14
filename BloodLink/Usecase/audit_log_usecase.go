package Usecase

import (
	"bloodlink/Domain"
	"context"
)

type IAuditLogRepository interface {
	CreateLog(ctx context.Context, audit *Domain.AuditLog) error
	GetLogs(ctx context.Context, filter Domain.AuditLogFilter) ([]Domain.AuditLog, error)
	GetLogByID(ctx context.Context, id int64) (*Domain.AuditLog, error)
	DeleteLog(ctx context.Context, id int64) error
}

type AuditLogUsecase struct {
	Repo IAuditLogRepository
}

func NewAuditLogUsecase(repo IAuditLogRepository) *AuditLogUsecase {
	return &AuditLogUsecase{Repo: repo}
}

func (u *AuditLogUsecase) Log(ctx context.Context, userID, action, targetType, targetID, targetName, oldValue, newValue string) {
	if oldValue == "" {
		oldValue = "N/A"
	}
	if newValue == "" {
		newValue = "N/A"
	}
	audit := &Domain.AuditLog{
		UserID:     userID,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		TargetName: targetName,
		OldValue:   oldValue,
		NewValue:   newValue,
	}
	_ = u.Repo.CreateLog(ctx, audit)
}

func (u *AuditLogUsecase) GetLogByID(ctx context.Context, id int64) (*Domain.AuditLog, error) {
	return u.Repo.GetLogByID(ctx, id)
}

func (u *AuditLogUsecase) GetLogs(ctx context.Context, filter Domain.AuditLogFilter) ([]Domain.AuditLog, error) {
	return u.Repo.GetLogs(ctx, filter)
}

func (u *AuditLogUsecase) DeleteLog(ctx context.Context, id int64) error {
	return u.Repo.DeleteLog(ctx, id)
}
