package Repository

import (
	domain "bloodlink/Domain"
	"context"
	"database/sql"
	"fmt"
)

type ProfileRepository struct {
	DB *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{DB: db}
}

func (r *ProfileRepository) CreateProfile(ctx context.Context, profile *domain.UserProfile) error {
	query := `INSERT INTO user_profiles (profile_id, user_id, full_name, phone, address, profile_picture_url, latitude, longitude, location_geo) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, ST_SetSRID(ST_MakePoint($8, $7), 4326)::geography)`

	_, err := r.DB.ExecContext(ctx, query,
		profile.ProfileID,
		profile.UserID,
		profile.FullName,
		profile.Phone,
		profile.Address,
		profile.ProfilePictureURL,
		profile.Latitude,
		profile.Longitude,
	)
	return err
}

func (r *ProfileRepository) GetProfileByUserID(ctx context.Context, userID string) (*domain.UserProfile, error) {
	query := `SELECT profile_id, user_id, COALESCE(full_name, ''), COALESCE(phone, ''), COALESCE(address, ''), COALESCE(profile_picture_url, ''), latitude, longitude FROM user_profiles WHERE user_id = $1`
	row := r.DB.QueryRowContext(ctx, query, userID)

	var profile domain.UserProfile
	err := row.Scan(
		&profile.ProfileID,
		&profile.UserID,
		&profile.FullName,
		&profile.Phone,
		&profile.Address,
		&profile.ProfilePictureURL,
		&profile.Latitude,
		&profile.Longitude,
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
	query := `UPDATE user_profiles SET full_name = $1, phone = $2, address = $3, profile_picture_url = $4, latitude = $5, longitude = $6, location_geo = ST_SetSRID(ST_MakePoint($6, $5), 4326)::geography WHERE user_id = $7`
	_, err := r.DB.ExecContext(ctx, query,
		profile.FullName,
		profile.Phone,
		profile.Address,
		profile.ProfilePictureURL,
		profile.Latitude,
		profile.Longitude,
		profile.UserID,
	)
	return err
}

func (r *ProfileRepository) GetAllProfiles(ctx context.Context, filter domain.ProfileFilter) ([]domain.UserProfile, error) {
	query := `
		SELECT 
			p.profile_id, 
			p.user_id, 
			COALESCE(p.full_name, ''), 
			COALESCE(p.phone, ''), 
			COALESCE(p.address, ''), 
			COALESCE(p.profile_picture_url, ''), 
			p.latitude, 
			p.longitude 
		FROM user_profiles p
		JOIN users u ON p.user_id = u.user_id
		WHERE 1=1
	`
	args := []interface{}{}
	placeholderID := 1

	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND u.created_at >= $%d", placeholderID)
		args = append(args, filter.StartDate)
		placeholderID++
	}

	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND u.created_at <= $%d", placeholderID)
		args = append(args, filter.EndDate)
		placeholderID++
	}

	query += " ORDER BY u.created_at DESC"

	rows, err := r.DB.QueryContext(ctx, query, args...)
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
			&profile.Latitude,
			&profile.Longitude,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}
