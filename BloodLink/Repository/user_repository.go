package Repository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	domain "bloodlink/Domain"
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
               VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
	query := `SELECT user_id, full_name, email, COALESCE(phone, ''), password_hash, role, is_active, COALESCE(otp, ''), created_at, COALESCE(refresh_token, '') FROM users WHERE email = ?`

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

// ActivateUser updates the user's status to active and clears the OTP
func (r *UserRepository) ActivateUser(ctx context.Context, userID string) error {
	query := `UPDATE users SET is_active = true, otp = NULL WHERE user_id = ?`
	_, err := r.DB.ExecContext(ctx, query, userID)
	return err
}

// CreateDonor inserts a minimal donor record into the database
func (r *UserRepository) CreateDonor(ctx context.Context, donor *domain.Donor) error {
	query := `INSERT INTO donors (donor_id, user_id, status) VALUES (?, ?, ?)`
	_, err := r.DB.ExecContext(ctx, query, donor.DonorID, donor.UserID, donor.Status)
	return err
}

// DeleteUser removes a user from the database.
// Due to ON DELETE CASCADE, this will also remove their Profile and Donor record.
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE user_id = ?`
	_, err := r.DB.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("[DATABASE ERROR] DeleteUser failed: %v", err)
		return err
	}
	return nil
}

// SetOTP stores an OTP for the user identified by email (used for forgot password)
func (r *UserRepository) SetOTP(ctx context.Context, email, otp string) error {
	query := `UPDATE users SET otp = ? WHERE email = ?`
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
	query := `UPDATE users SET password_hash = ?, otp = NULL WHERE email = ?`
	_, err := r.DB.ExecContext(ctx, query, hashedPassword, email)
	if err != nil {
		log.Printf("[DATABASE ERROR] ResetPassword failed: %v", err)
		return err
	}
	return nil
}

// UpdateDonorStatus updates the status of a donor by donor_id
func (r *UserRepository) UpdateDonorStatus(ctx context.Context, donorID, status string) error {
	query := `UPDATE donors SET status = ? WHERE donor_id = ?`
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
			d.status 
		FROM donors d
		JOIN users u ON d.user_id = u.user_id
		LEFT JOIN user_profile p ON u.user_id = p.user_id
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
			d.status 
		FROM donors d
		JOIN users u ON d.user_id = u.user_id
		LEFT JOIN user_profile p ON u.user_id = p.user_id
		WHERE 1=1
	`
	args := []interface{}{}

	if filter.BloodType != "" {
		query += " AND d.blood_type = ?"
		args = append(args, filter.BloodType)
	}
	if filter.Status != "" {
		query += " AND d.status = ?"
		args = append(args, filter.Status)
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
		); err != nil {
			return nil, err
		}
		donors = append(donors, donor)
	}

	return donors, nil
}

// GetUsersByRole retrieves all users matching a specific role
func (r *UserRepository) GetUsersByRole(ctx context.Context, role string) ([]domain.UserResponse, error) {
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
		WHERE role = ?
	`
	rows, err := r.DB.QueryContext(ctx, query, role)
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
	query := `UPDATE users SET refresh_token = ? WHERE user_id = ?`
	_, err := r.DB.ExecContext(ctx, query, refreshToken, userID)
	return err
}
