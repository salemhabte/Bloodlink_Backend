package Repository

import (
	domain "bloodlink/Domain"
	"context"
	"database/sql"
)

type ProfileRepository struct {
	DB *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{DB: db}
}

func (r *ProfileRepository) CreateProfile(ctx context.Context, profile *domain.UserProfile) error {
	query := `INSERT INTO user_profiles (profile_id, user_id, full_name, phone, address, profile_picture_url) 
              VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.DB.ExecContext(ctx, query,
		profile.ProfileID,
		profile.UserID,
		profile.FullName,
		profile.Phone,
		profile.Address,
		profile.ProfilePictureURL,
	)
	return err
}

func (r *ProfileRepository) GetProfileByUserID(ctx context.Context, userID string) (*domain.UserProfile, error) {
	query := `SELECT profile_id, user_id, COALESCE(full_name, ''), COALESCE(phone, ''), COALESCE(address, ''), COALESCE(profile_picture_url, '') FROM user_profiles WHERE user_id = ?`
	row := r.DB.QueryRowContext(ctx, query, userID)

	var profile domain.UserProfile
	err := row.Scan(
		&profile.ProfileID,
		&profile.UserID,
		&profile.FullName,
		&profile.Phone,
		&profile.Address,
		&profile.ProfilePictureURL,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

func (r *ProfileRepository) UpdateProfile(ctx context.Context, profile *domain.UserProfile) error {
	query := `UPDATE user_profiles SET full_name = ?, phone = ?, address = ?, profile_picture_url = ? WHERE user_id = ?`
	_, err := r.DB.ExecContext(ctx, query,
		profile.FullName,
		profile.Phone,
		profile.Address,
		profile.ProfilePictureURL,
		profile.UserID,
	)
	return err
}

func (r *ProfileRepository) GetAllProfiles(ctx context.Context) ([]domain.UserProfile, error) {
	query := `SELECT profile_id, user_id, COALESCE(full_name, ''), COALESCE(phone, ''), COALESCE(address, ''), COALESCE(profile_picture_url, '') FROM user_profiles`
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []domain.UserProfile
	for rows.Next() {
		var profile domain.UserProfile
		if err := rows.Scan(
			&profile.ProfileID,
			&profile.UserID,
			&profile.FullName,
			&profile.Phone,
			&profile.Address,
			&profile.ProfilePictureURL,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}
