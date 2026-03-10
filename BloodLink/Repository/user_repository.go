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
	query := `INSERT INTO Users (user_id, email, full_name, phone, password_hash, role, is_active, otp, created_at) 
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
	query := `SELECT user_id, full_name, email, phone, password_hash, role, is_active, otp, created_at FROM Users WHERE email = ?`

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
	query := `UPDATE Users SET is_active = true, otp = NULL WHERE user_id = ?`
	_, err := r.DB.ExecContext(ctx, query, userID)
	return err
}

// CreateDonor inserts a minimal donor record into the database
func (r *UserRepository) CreateDonor(ctx context.Context, donor *domain.Donor) error {
	query := `INSERT INTO Donors (donor_id, user_id, status) VALUES (?, ?, ?)`
	_, err := r.DB.ExecContext(ctx, query, donor.DonorID, donor.UserID, donor.Status)
	return err
}

// DeleteUser removes a user from the database.
// Due to ON DELETE CASCADE, this will also remove their Profile and Donor record.
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	query := `DELETE FROM Users WHERE user_id = ?`
	_, err := r.DB.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("[DATABASE ERROR] DeleteUser failed: %v", err)
		return err
	}
	return nil
}

