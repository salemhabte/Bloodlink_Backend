package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AdminAuditLogRepository struct {
	DB *sql.DB
}

func NewAdminAuditLogRepository(db *sql.DB) Interfaces.IAdminAuditLogRepository {
	return &AdminAuditLogRepository{DB: db}
}

func (r *AdminAuditLogRepository) CreateLog(log *Domain.AdminAuditLog) error {
	log.LogID = uuid.New().String()
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	query := `
	INSERT INTO admin_audit_logs 
	(log_id, admin_id, action, target_type, target_id, details, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.DB.Exec(query,
		log.LogID,
		log.AdminID,
		log.Action,
		log.TargetType,
		log.TargetID,
		log.Details,
		log.CreatedAt,
	)

	return err
}

func (r *AdminAuditLogRepository) GetLogs(filter Domain.AuditLogFilter) ([]Domain.AdminAuditLog, int, error) {
	// Base query
	query := `
	SELECT log_id, admin_id, action, target_type, target_id, details, created_at
	FROM admin_audit_logs
	WHERE 1=1
	`
	countQuery := `
	SELECT COUNT(*)
	FROM admin_audit_logs
	WHERE 1=1
	`

	args := []interface{}{}
	placeholderID := 1

	if filter.Action != "" {
		query += fmt.Sprintf(" AND action = $%d", placeholderID)
		countQuery += fmt.Sprintf(" AND action = $%d", placeholderID)
		args = append(args, filter.Action)
		placeholderID++
	}

	if filter.TargetType != "" {
		query += fmt.Sprintf(" AND target_type = $%d", placeholderID)
		countQuery += fmt.Sprintf(" AND target_type = $%d", placeholderID)
		args = append(args, filter.TargetType)
		placeholderID++
	}

	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND created_at >= $%d", placeholderID)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", placeholderID)
		args = append(args, filter.StartDate)
		placeholderID++
	}

	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND created_at <= $%d", placeholderID)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", placeholderID)
		args = append(args, filter.EndDate)
		placeholderID++
	}

	// Count total
	var total int
	err := r.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}
	offset := (filter.Page - 1) * filter.Limit

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", placeholderID, placeholderID+1)
	args = append(args, filter.Limit, offset)

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []Domain.AdminAuditLog
	for rows.Next() {
		var l Domain.AdminAuditLog
		err := rows.Scan(
			&l.LogID,
			&l.AdminID,
			&l.Action,
			&l.TargetType,
			&l.TargetID,
			&l.Details,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}

	return logs, total, nil
}

func (r *AdminAuditLogRepository) GetLogByID(id string) (*Domain.AdminAuditLog, error) {
	query := `
	SELECT log_id, admin_id, action, target_type, target_id, details, created_at
	FROM admin_audit_logs
	WHERE log_id = $1
	`
	var l Domain.AdminAuditLog
	err := r.DB.QueryRow(query, id).Scan(
		&l.LogID,
		&l.AdminID,
		&l.Action,
		&l.TargetType,
		&l.TargetID,
		&l.Details,
		&l.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audit log not found")
		}
		return nil, err
	}
	return &l, nil
}

func (r *AdminAuditLogRepository) DeleteLog(id string) error {
	query := `DELETE FROM admin_audit_logs WHERE log_id = $1`
	result, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("audit log not found")
	}
	
	return nil
}
