package Repository

import (
	"context"
	"database/sql"
	domain "bloodlink/Domain"
)

type ProfileRepository struct {
	DB *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{DB: db}
}

func (r *ProfileRepository) CreateProfile(ctx context.Context, profile *domain.UserProfile) error {
	query := `INSERT INTO User_Profile (profile_id, user_id, full_name, phone, city, area, profile_picture_url) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`
			  
	_, err := r.DB.ExecContext(ctx, query,
		profile.ProfileID,
		profile.UserID,
		profile.FullName,
		profile.Phone,
		profile.City,
		profile.Area,
		profile.ProfilePictureURL,
	)
	
	return err
}

func (r *ProfileRepository) GetProfileByUserID(ctx context.Context, userID string) (*domain.UserProfile, error) {
	query := `SELECT profile_id, user_id, full_name, phone, city, area, profile_picture_url FROM User_Profile WHERE user_id = ?`
	row := r.DB.QueryRowContext(ctx, query, userID)

	var profile domain.UserProfile
	err := row.Scan(
		&profile.ProfileID,
		&profile.UserID,
		&profile.FullName,
		&profile.Phone,
		&profile.City,
		&profile.Area,
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
	query := `UPDATE User_Profile SET full_name = ?, phone = ?, city = ?, area = ?, profile_picture_url = ? WHERE user_id = ?`
	_, err := r.DB.ExecContext(ctx, query,
		profile.FullName,
		profile.Phone,
		profile.City,
		profile.Area,
		profile.ProfilePictureURL,
		profile.UserID,
	)
	return err
}
