package Repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	domain "bloodlink/Domain"
	"strings"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser inserts a newly registered user into the database
func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (user_id, email, full_name, phone, password_hash, role, is_active, otp, created_at) 
               VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	// Note: ER diagram says password_hash, domain says password.
	// We'll map the db columns according to ER diagram or logic.
	_, err := r.DB.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.FullName,
		user.Phone,
		user.Password,
		user.Role,
		user.IsActive,
		user.OTP,
		user.CreatedAt,
	)

	if err != nil {
		log.Printf("[DATABASE ERROR] CreateUser failed: %v", err)
		return err
	}

	return nil
}

// GetUserByEmail retrieves a user by their email address for login verification
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT user_id, full_name, email, COALESCE(phone, ''), password_hash, role, is_active, COALESCE(otp, ''), created_at, COALESCE(refresh_token, '') FROM users WHERE email = $1`

	row := r.DB.QueryRowContext(ctx, query, email)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.OTP,
		&user.CreatedAt,
		&user.RefreshToken,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil, nil when no user is found
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByPhone retrieves a user by their phone number
func (r *UserRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `SELECT user_id, full_name, email, phone, password_hash, role, is_active, COALESCE(otp, ''), created_at FROM users WHERE phone = $1`

	var user domain.User
	err := r.DB.QueryRowContext(ctx, query, phone).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.Role,
		&user.IsActive,
		&user.OTP,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// ActivateUser updates the user's status to active and clears the OTP
func (r *UserRepository) ActivateUser(ctx context.Context, userID string) error {
	query := `UPDATE users SET is_active = true, otp = NULL WHERE user_id = $1`
	_, err := r.DB.ExecContext(ctx, query, userID)
	return err
}

// CreateDonor inserts a donor record into the database
func (r *UserRepository) CreateDonor(ctx context.Context, donor *domain.Donor) error {
	query := `INSERT INTO donors (donor_id, user_id, overall_status, date_of_birth) VALUES ($1, $2, $3, $4)`
	_, err := r.DB.ExecContext(ctx, query, donor.DonorID, donor.UserID, donor.OverallStatus, donor.DateOfBirth)
	return err
}

func (r *UserRepository) GetDonorByUserID(ctx context.Context, userID string) (*domain.Donor, error) {
	query := `SELECT donor_id, user_id, overall_status, date_of_birth, blood_type FROM donors WHERE user_id = $1`
	donor := &domain.Donor{}
	var dob sql.NullTime
	var bt sql.NullString

	err := r.DB.QueryRowContext(ctx, query, userID).Scan(&donor.DonorID, &donor.UserID, &donor.OverallStatus, &dob, &bt)
	if err != nil {
		return nil, err
	}

	if dob.Valid {
		donor.DateOfBirth = dob.Time
	}
	if bt.Valid {
		donor.BloodType = bt.String
	}

	return donor, nil
}

// DeleteUser removes a user from the database.
// Due to ON DELETE CASCADE, this will also remove their Profile and Donor record.
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE user_id = $1`
	_, err := r.DB.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("[DATABASE ERROR] DeleteUser failed: %v", err)
		return err
	}
	return nil
}

// SetOTP stores an OTP for the user identified by email (used for forgot password)
func (r *UserRepository) SetOTP(ctx context.Context, email, otp string) error {
	query := `UPDATE users SET otp = $1 WHERE email = $2`
	result, err := r.DB.ExecContext(ctx, query, otp, email)
	if err != nil {
		log.Printf("[DATABASE ERROR] SetOTP failed: %v", err)
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("user not found")
	}
	return nil
}

// ResetPassword updates the password and clears the OTP for the user identified by email
func (r *UserRepository) ResetPassword(ctx context.Context, email, hashedPassword string) error {
	query := `UPDATE users SET password_hash = $1, otp = NULL WHERE email = $2`
	_, err := r.DB.ExecContext(ctx, query, hashedPassword, email)
	if err != nil {
		log.Printf("[DATABASE ERROR] ResetPassword failed: %v", err)
		return err
	}
	return nil
}

// UpdateDonorStatus updates the lab/final screening status of a donor
func (r *UserRepository) UpdateDonorStatus(ctx context.Context, donorID, status string) error {
	query := `UPDATE donors SET overall_status = $1 WHERE donor_id = $2`
	result, err := r.DB.ExecContext(ctx, query, status, donorID)
	if err != nil {
		log.Printf("[DATABASE ERROR] UpdateDonorStatus failed: %v", err)
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("donor not found")
	}
	return nil
}

func (r *UserRepository) GetAllDonors(ctx context.Context) ([]domain.DonorResponse, error) {
	query := `
		SELECT 
			d.donor_id, 
			d.user_id, 
			u.full_name, 
			u.email, 
			u.phone, 
			COALESCE(p.address, ''), 
			COALESCE(d.blood_type, ''), 
			COALESCE(dr.last_status, 'Pending'),
			d.overall_status
		FROM donors d
		JOIN users u ON d.user_id = u.user_id
		LEFT JOIN user_profiles p ON u.user_id = p.user_id
		LEFT JOIN (
			SELECT DISTINCT ON (donor_id) donor_id, status as last_status
			FROM donation_records
			ORDER BY donor_id, collection_date DESC, created_at DESC
		) dr ON d.donor_id = dr.donor_id
	`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		log.Printf("[DATABASE ERROR] GetAllDonors failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var donors []domain.DonorResponse
	for rows.Next() {
		var donor domain.DonorResponse
		if err := rows.Scan(
			&donor.DonorID,
			&donor.UserID,
			&donor.FullName,
			&donor.Email,
			&donor.Phone,
			&donor.Address,
			&donor.BloodType,
			&donor.Status,
			&donor.OverallStatus,
		); err != nil {
			return nil, err
		}
		donors = append(donors, donor)
	}

	return donors, nil
}

func (r *UserRepository) FilterDonors(ctx context.Context, filter domain.DonorFilter) ([]domain.DonorResponse, error) {
	query := `
		SELECT 
			d.donor_id, 
			d.user_id, 
			u.full_name, 
			u.email, 
			u.phone, 
			COALESCE(p.address, ''), 
			COALESCE(d.blood_type, ''), 
			COALESCE(dr.last_status, 'Pending'),
			d.overall_status,
			u.created_at
		FROM donors d
		JOIN users u ON d.user_id = u.user_id
		LEFT JOIN user_profiles p ON u.user_id = p.user_id
		LEFT JOIN (
			SELECT DISTINCT ON (donor_id) donor_id, status as last_status, collection_date as last_donation
			FROM donation_records
			ORDER BY donor_id, collection_date DESC, created_at DESC
		) dr ON d.donor_id = dr.donor_id
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.BloodType != "" {
		args = append(args, filter.BloodType)
		query += fmt.Sprintf(" AND d.blood_type = $%d", len(args))
	}
	if filter.OverallStatus != "" {
		s := strings.ToLower(filter.OverallStatus)
		if s == "cleared" {
			query += " AND d.overall_status = 'CLEARED'"
		} else if s == "temporarily_deferred" {
			query += " AND d.overall_status = 'TEMPORARILY_DEFERRED'"
		} else if s == "permanently_deferred" {
			query += " AND d.overall_status = 'PERMANENTLY_DEFERRED'"
		} else {
			args = append(args, filter.OverallStatus)
			query += fmt.Sprintf(" AND d.overall_status = $%d", len(args))
		}
	}
	if filter.StartDate != "" {
		args = append(args, filter.StartDate)
		query += fmt.Sprintf(" AND u.created_at >= $%d", len(args))
	}
	if filter.EndDate != "" {
		args = append(args, filter.EndDate)
		query += fmt.Sprintf(" AND u.created_at <= $%d", len(args))
	}
	if filter.Status != "" {
		s := strings.ToLower(filter.Status)
		if s == "active" || s == "inactive" {
			isActive := (s == "active")
			args = append(args, isActive)
			query += fmt.Sprintf(" AND u.is_active = $%d", len(args))
		} else if s == "approved" {
			query += " AND dr.last_status = 'APPROVED'"
		} else if s == "rejected_temporary" || s == "temporarily_rejected" {
			query += " AND dr.last_status = 'REJECTED_TEMPORARY'"
		} else if s == "pending" {
			query += " AND dr.last_status = 'PENDING'"
		}
	}
	if filter.IsNewDonor != nil {
		if *filter.IsNewDonor {
			query += " AND dr.last_donation IS NULL"
		} else {
			query += " AND dr.last_donation IS NOT NULL"
		}
	}
	if filter.IsEligible != nil {
		if *filter.IsEligible {
			query += " AND (dr.last_donation IS NULL OR (dr.last_donation <= NOW() - INTERVAL '90 days' AND d.overall_status != 'Pending')) AND d.overall_status != 'PERMANENTLY_DEFERRED'"
		} else {
			query += " AND ((dr.last_donation IS NOT NULL AND (dr.last_donation > NOW() - INTERVAL '90 days' OR d.overall_status = 'Pending')) OR d.overall_status = 'PERMANENTLY_DEFERRED')"
		}
	}

	query += " ORDER BY u.created_at DESC"
	if filter.StartDate != "" {
		args = append(args, filter.StartDate)
		query += fmt.Sprintf(" AND u.created_at >= $%d", len(args))
	}
	if filter.EndDate != "" {
		args = append(args, filter.EndDate)
		query += fmt.Sprintf(" AND u.created_at <= $%d", len(args))
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("[DATABASE ERROR] FilterDonors failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var donors []domain.DonorResponse
	for rows.Next() {
		var donor domain.DonorResponse
		if err := rows.Scan(
			&donor.DonorID,
			&donor.UserID,
			&donor.FullName,
			&donor.Email,
			&donor.Phone,
			&donor.Address,
			&donor.BloodType,
			&donor.Status,
			&donor.OverallStatus,
			&donor.RegistrationDate,
		); err != nil {
			return nil, err
		}
		donors = append(donors, donor)
	}

	return donors, nil
}

// GetUsersByRole retrieves all users matching a specific role (or all users if role is empty)
func (r *UserRepository) GetUsersByRole(ctx context.Context, filter domain.UserFilter) ([]domain.UserResponse, error) {
	query := `
		SELECT 
			user_id, 
			full_name, 
			email, 
			COALESCE(phone, ''), 
			role, 
			is_active, 
			created_at 
		FROM users 
		WHERE 1=1
	`
	args := []interface{}{}
	placeholderID := 1

	if filter.Role != "" {
		query += fmt.Sprintf(" AND role = $%d", placeholderID)
		args = append(args, filter.Role)
		placeholderID++
	}

	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND created_at >= $%d", placeholderID)
		args = append(args, filter.StartDate)
		placeholderID++
	}

	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND created_at <= $%d", placeholderID)
		args = append(args, filter.EndDate)
		placeholderID++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("[DATABASE ERROR] GetUsersByRole failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var users []domain.UserResponse
	for rows.Next() {
		var u domain.UserResponse
		if err := rows.Scan(
			&u.ID,
			&u.FullName,
			&u.Email,
			&u.Phone,
			&u.Role,
			&u.IsActive,
			&u.CreatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) UpdateRefreshToken(ctx context.Context, userID, refreshToken string) error {
	query := `UPDATE users SET refresh_token = $1 WHERE user_id = $2`
	_, err := r.DB.ExecContext(ctx, query, refreshToken, userID)
	return err
}

func (r *UserRepository) GetDonorsByBloodTypeAndAddress(ctx context.Context, bloodType, address string) ([]domain.DonorResponse, error) {
	query := `
		SELECT 
			d.donor_id, 
			d.user_id, 
			u.full_name, 
			u.email, 
			u.phone, 
			COALESCE(p.address, ''), 
			COALESCE(d.blood_type, ''), 
			COALESCE(dr.last_status, 'Pending'),
			d.overall_status 
		FROM donors d
		JOIN users u ON d.user_id = u.user_id
		LEFT JOIN user_profiles p ON u.user_id = p.user_id
		LEFT JOIN (
			SELECT DISTINCT ON (donor_id) donor_id, status as last_status
			FROM donation_records
			ORDER BY donor_id, collection_date DESC, created_at DESC
		) dr ON d.donor_id = dr.donor_id
		WHERE d.blood_type = $1 AND p.address ILIKE $2
	`
	rows, err := r.DB.QueryContext(ctx, query, bloodType, "%"+address+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var donors []domain.DonorResponse
	for rows.Next() {
		var donor domain.DonorResponse
		if err := rows.Scan(
			&donor.DonorID,
			&donor.UserID,
			&donor.FullName,
			&donor.Email,
			&donor.Phone,
			&donor.Address,
			&donor.BloodType,
			&donor.Status,
			&donor.OverallStatus,
		); err != nil {
			return nil, err
		}
		donors = append(donors, donor)
	}
	return donors, nil
}

func (r *UserRepository) GetEligibleDonors(ctx context.Context, searchQuery string) ([]domain.DonorResponse, error) {
	query := `
		SELECT 
			d.donor_id, 
			d.user_id, 
			u.full_name, 
			u.email, 
			u.phone, 
			COALESCE(p.address, ''), 
			COALESCE(d.blood_type, ''), 
			COALESCE(dr.last_status, 'Pending'),
			d.overall_status 
		FROM donors d
		JOIN users u ON d.user_id = u.user_id
		LEFT JOIN user_profiles p ON u.user_id = p.user_id
		LEFT JOIN (
			SELECT DISTINCT ON (donor_id) donor_id, status as last_status, collection_date as last_donation
			FROM donation_records
			ORDER BY donor_id, collection_date DESC, created_at DESC
		) dr ON d.donor_id = dr.donor_id
		WHERE (dr.last_donation IS NULL OR (dr.last_donation <= NOW() - INTERVAL '90 days' AND d.overall_status != 'Pending'))
		AND d.overall_status != 'PERMANENTLY_DEFERRED'
	`
	args := []interface{}{}
	if searchQuery != "" {
		searchTerm := "%" + searchQuery + "%"
		args = append(args, searchTerm)
		query += fmt.Sprintf(" AND (u.email ILIKE $%d OR u.phone LIKE $%d OR u.full_name ILIKE $%d)", len(args), len(args), len(args))
	}

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("[DATABASE ERROR] GetEligibleDonors failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var donors []domain.DonorResponse
	for rows.Next() {
		var donor domain.DonorResponse
		if err := rows.Scan(
			&donor.DonorID,
			&donor.UserID,
			&donor.FullName,
			&donor.Email,
			&donor.Phone,
			&donor.Address,
			&donor.BloodType,
			&donor.Status,
			&donor.OverallStatus,
		); err != nil {
			return nil, err
		}
		donors = append(donors, donor)
	}

	return donors, nil
}

func (r *UserRepository) GetDonorsBecomingEligibleToday(ctx context.Context) ([]domain.DonorResponse, error) {
	query := `
		SELECT 
			u.user_id, 
			u.full_name, 
			u.email, 
			u.phone
		FROM donors d
		JOIN users u ON d.user_id = u.user_id
		JOIN (
			SELECT donor_id, MAX(collection_date) as last_donation
			FROM donation_records
			GROUP BY donor_id
		) dr ON d.donor_id = dr.donor_id
		WHERE dr.last_donation = CURRENT_DATE - INTERVAL '90 days'
		AND d.overall_status != 'PERMANENTLY_DEFERRED'
	`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var donors []domain.DonorResponse
	for rows.Next() {
		var d domain.DonorResponse
		if err := rows.Scan(&d.UserID, &d.FullName, &d.Email, &d.Phone); err != nil {
			return nil, err
		}
		donors = append(donors, d)
	}
	return donors, nil
}

func (r *UserRepository) GetDonorsNearby(ctx context.Context, bloodType string, lat, lon, radiusKm float64) ([]domain.DonorResponse, error) {
	query := `
		SELECT 
			d.donor_id, 
			d.user_id, 
			u.full_name, 
			u.email, 
			u.phone, 
			COALESCE(p.address, ''), 
			COALESCE(d.blood_type, ''), 
			d.overall_status 
		FROM donors d
		JOIN users u ON d.user_id = u.user_id
		LEFT JOIN user_profiles p ON u.user_id = p.user_id
		WHERE d.blood_type = $1
		AND p.location_geo IS NOT NULL
		AND ST_DWithin(p.location_geo, ST_SetSRID(ST_MakePoint($3, $2), 4326)::geography, $4 * 1000)
		AND d.overall_status != 'PERMANENTLY_DEFERRED'
		ORDER BY ST_Distance(p.location_geo, ST_SetSRID(ST_MakePoint($3, $2), 4326)::geography) ASC
	`
	rows, err := r.DB.QueryContext(ctx, query, bloodType, lat, lon, radiusKm)
	if err != nil {
		log.Printf("[DATABASE ERROR] GetDonorsNearby failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	var donors []domain.DonorResponse
	for rows.Next() {
		var donor domain.DonorResponse
		if err := rows.Scan(
			&donor.DonorID,
			&donor.UserID,
			&donor.FullName,
			&donor.Email,
			&donor.Phone,
			&donor.Address,
			&donor.BloodType,
			&donor.OverallStatus,
		); err != nil {
			return nil, err
		}
		donors = append(donors, donor)
	}
	return donors, nil
}
