package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CampaignRepository implements Domain.ICampaignRepository
type CampaignRepository struct {
	DB *sql.DB
}

// NewCampaignRepository creates a new repository instance
func NewCampaignRepository(db *sql.DB) Interfaces.ICampaignRepository {
	return &CampaignRepository{DB: db}
}

// CreateCampaign inserts a new campaign
func (r *CampaignRepository) CreateCampaign(campaign *Domain.Campaign) error {
	campaign.CampaignID = uuid.New().String()
	campaign.CreatedAt = time.Now()

	query := `
	INSERT INTO campaigns 
	(campaign_id, title, content, location, start_date, end_date, created_at, status) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.DB.Exec(
		query,
		campaign.CampaignID,
		campaign.Title,
		campaign.Content,
		campaign.Location,
		campaign.StartDate,
		campaign.EndDate,
		campaign.CreatedAt,
		campaign.Status,
	)

	return err
}

func (r *CampaignRepository) GetAllCampaigns(filter Domain.CampaignFilter, liveOnly bool) ([]Domain.Campaign, error) {
	query := `
	SELECT campaign_id, title, content, location, start_date, end_date, created_at, is_deleted, status
	FROM campaigns
	WHERE is_deleted = false
	`
	args := []interface{}{}
	placeholderID := 1

	if liveOnly || filter.LiveOnly {
		query += fmt.Sprintf(" AND end_date >= $%d", placeholderID)
		args = append(args, time.Now())
		placeholderID++
	}

	if filter.Title != "" {
		query += fmt.Sprintf(" AND title ILIKE $%d", placeholderID)
		args = append(args, "%"+filter.Title+"%")
		placeholderID++
	}

	if filter.Location != "" {
		query += fmt.Sprintf(" AND location ILIKE $%d", placeholderID)
		args = append(args, "%"+filter.Location+"%")
		placeholderID++
	}

	if filter.StartDate != "" {
		query += fmt.Sprintf(" AND start_date >= $%d", placeholderID)
		args = append(args, filter.StartDate)
		placeholderID++
	}

	if filter.EndDate != "" {
		query += fmt.Sprintf(" AND end_date <= $%d", placeholderID)
		args = append(args, filter.EndDate)
		placeholderID++
	}

	// SORT: Ongoing first, then Upcoming, then Past. Within groups, sort by end_date ASC.
	query += ` ORDER BY (
		CASE 
			WHEN end_date >= NOW() AND start_date <= NOW() THEN 0 
			WHEN end_date >= NOW() AND start_date > NOW() THEN 1 
			ELSE 2 
		END), end_date ASC`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []Domain.Campaign

	for rows.Next() {
		var c Domain.Campaign

		err := rows.Scan(
			&c.CampaignID,
			&c.Title,
			&c.Content,
			&c.Location,
			&c.StartDate,
			&c.EndDate,
			&c.CreatedAt,
			&c.IsDeleted,
			&c.Status,
		)

		if err != nil {
			return nil, err
		}

		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}

// GetCampaignByID returns a campaign by ID (General)
func (r *CampaignRepository) GetCampaignByID(id string) (*Domain.Campaign, error) {

	query := `
	SELECT campaign_id, title, content, location, start_date, end_date, created_at, is_deleted, status
	FROM campaigns
	WHERE campaign_id = $1 AND is_deleted = false
	LIMIT 1
	`

	row := r.DB.QueryRow(query, id)

	var c Domain.Campaign

	err := row.Scan(
		&c.CampaignID,
		&c.Title,
		&c.Content,
		&c.Location,
		&c.StartDate,
		&c.EndDate,
		&c.CreatedAt,
		&c.IsDeleted,
		&c.Status,
	)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

// GetLiveCampaignByID returns a campaign by ID ONLY if it is not expired
func (r *CampaignRepository) GetLiveCampaignByID(id string) (*Domain.Campaign, error) {

	query := `
	SELECT campaign_id, title, content, location, start_date, end_date, created_at, is_deleted, status
	FROM campaigns
	WHERE campaign_id = $1 AND start_date <= NOW() AND end_date >= NOW() AND is_deleted = false
	LIMIT 1
	`

	row := r.DB.QueryRow(query, id)

	var c Domain.Campaign

	err := row.Scan(
		&c.CampaignID,
		&c.Title,
		&c.Content,
		&c.Location,
		&c.StartDate,
		&c.EndDate,
		&c.CreatedAt,
		&c.IsDeleted,
		&c.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("campaign is not currently active (it may have ended or not yet started)")
		}
		return nil, err
	}

	return &c, nil
}

// UpdateCampaign updates an existing campaign
func (r *CampaignRepository) UpdateCampaign(campaign *Domain.Campaign) error {

	existing, err := r.GetCampaignByID(campaign.CampaignID)
	if err != nil {
		return err
	}

	if campaign.Title != "" {
		existing.Title = campaign.Title
	}

	if campaign.Content != "" {
		existing.Content = campaign.Content
	}

	if campaign.Location != "" {
		existing.Location = campaign.Location
	}

	if !campaign.StartDate.IsZero() {
		existing.StartDate = campaign.StartDate
	}

	if !campaign.EndDate.IsZero() {
		existing.EndDate = campaign.EndDate
	}

	query := `
	UPDATE campaigns
	SET title=$1, content=$2, location=$3, start_date=$4, end_date=$5, status=$6
	WHERE campaign_id=$7
	`

	_, err = r.DB.Exec(
		query,
		existing.Title,
		existing.Content,
		existing.Location,
		existing.StartDate,
		existing.EndDate,
		existing.Status,
		existing.CampaignID,
	)

	return err
}

// DeleteCampaign removes a campaign
func (r *CampaignRepository) DeleteCampaign(id string) error {

	query := "UPDATE campaigns SET is_deleted = true WHERE campaign_id=$1"

	_, err := r.DB.Exec(query, id)

	return err
}

func (r *CampaignRepository) MarkClosedCampaigns() error {
	query := `UPDATE campaigns SET status = 'CLOSED' WHERE status != 'CLOSED' AND end_date <= $1`
	_, err := r.DB.Exec(query, time.Now())
	return err
}
