package Repository

import (
	"bloodlink/Domain"
	"context"
	"database/sql"
	"fmt"
	"log"
)

type AuditLogRepository struct {
	DB *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	return &AuditLogRepository{DB: db}
}

func (r *AuditLogRepository) CreateLog(ctx context.Context, audit *Domain.AuditLog) error {
	query := `INSERT INTO audit_logs (user_id, action, target_type, target_id, target_name, old_value, new_value) 
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.DB.ExecContext(ctx, query, 
		audit.UserID, 
		audit.Action, 
		audit.TargetType, 
		audit.TargetID, 
		audit.TargetName,
		audit.OldValue, 
		audit.NewValue,
	)
	if err != nil {
		log.Printf("[DATABASE ERROR] CreateLog failed: %v", err)
	}
	return err
}

func (r *AuditLogRepository) GetLogs(ctx context.Context, filter Domain.AuditLogFilter) ([]Domain.AuditLog, error) {
	query := `
		SELECT 
			a.log_id, a.user_id, 'Blood bank admin', a.action, a.target_type, a.target_id, a.target_name,
			COALESCE(NULLIF(a.old_value, ''), 'N/A'), COALESCE(NULLIF(a.new_value, ''), 'N/A'), a.created_at
		FROM audit_logs a
		WHERE a.deleted_at IS NULL
	`
	args := []interface{}{}

	if filter.UserID != "" {
		args = append(args, filter.UserID)
		query += fmt.Sprintf(" AND a.user_id = $%d", len(args))
	}
	if filter.Action != "" {
		args = append(args, filter.Action)
		query += fmt.Sprintf(" AND a.action = $%d", len(args))
	}
	if filter.TargetType != "" {
		args = append(args, filter.TargetType)
		query += fmt.Sprintf(" AND a.target_type = $%d", len(args))
	}
	if filter.StartDate != "" {
		args = append(args, filter.StartDate)
		query += fmt.Sprintf(" AND a.created_at >= $%d", len(args))
	}
	if filter.EndDate != "" {
		args = append(args, filter.EndDate)
		query += fmt.Sprintf(" AND a.created_at <= $%d", len(args))
	}

	query += " ORDER BY a.created_at DESC"

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]Domain.AuditLog, 0)
	for rows.Next() {
		var a Domain.AuditLog
		if err := rows.Scan(
			&a.LogID, &a.UserID, &a.UserName, &a.Action, &a.TargetType, &a.TargetID, &a.TargetName,
			&a.OldValue, &a.NewValue, &a.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, a)
	}
	return logs, nil
}
// GetLogByID retrieves a single audit log by its ID.
func (r *AuditLogRepository) GetLogByID(ctx context.Context, id int64) (*Domain.AuditLog, error) {
    query := `SELECT a.log_id, a.user_id, 'Blood bank admin', a.action, a.target_type, a.target_id, a.target_name,
        COALESCE(NULLIF(a.old_value, ''), 'N/A'), COALESCE(NULLIF(a.new_value, ''), 'N/A'), a.created_at
        FROM audit_logs a
        WHERE a.log_id = $1 AND a.deleted_at IS NULL`
    row := r.DB.QueryRowContext(ctx, query, id)
    var a Domain.AuditLog
    if err := row.Scan(&a.LogID, &a.UserID, &a.UserName, &a.Action, &a.TargetType, &a.TargetID, &a.TargetName, &a.OldValue, &a.NewValue, &a.CreatedAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &a, nil
}

// DeleteLog removes an audit log entry by its ID.
func (r *AuditLogRepository) DeleteLog(ctx context.Context, id int64) error {
    query := `UPDATE audit_logs SET deleted_at = CURRENT_TIMESTAMP WHERE log_id = $1`
    _, err := r.DB.ExecContext(ctx, query, id)
    return err
}

