package Repository

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"database/sql"
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
	(campaign_id, title, content, location, start_date, end_date, created_at) 
	VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.DB.Exec(
		query,
		campaign.CampaignID,
		campaign.Title,
		campaign.Content,
		campaign.Location,
		campaign.StartDate,
		campaign.EndDate,
		campaign.CreatedAt,
	)

	return err
}

// GetAllCampaigns returns all active campaigns
func (r *CampaignRepository) GetAllCampaigns() ([]Domain.Campaign, error) {

	query := `
	SELECT campaign_id, title, content, location, start_date, end_date, created_at
	FROM campaigns
	WHERE end_date >= NOW()
	`

	rows, err := r.DB.Query(query)
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
		)

		if err != nil {
			return nil, err
		}

		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}

// GetCampaignByID returns a campaign by ID
func (r *CampaignRepository) GetCampaignByID(id string) (*Domain.Campaign, error) {

	query := `
	SELECT campaign_id, title, content, location, start_date, end_date, created_at
	FROM campaigns
	WHERE campaign_id = $1 AND end_date >= NOW()
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
	)

	if err != nil {
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
	SET title=$1, content=$2, location=$3, start_date=$4, end_date=$5
	WHERE campaign_id=$6
	`

	_, err = r.DB.Exec(
		query,
		existing.Title,
		existing.Content,
		existing.Location,
		existing.StartDate,
		existing.EndDate,
		existing.CampaignID,
	)

	return err
}

// DeleteCampaign removes a campaign
func (r *CampaignRepository) DeleteCampaign(id string) error {

	query := "DELETE FROM campaigns WHERE campaign_id=$1"

	_, err := r.DB.Exec(query, id)

	return err
}

// GetCampaignsByLocation finds campaigns by location
func (r *CampaignRepository) GetCampaignsByLocation(location string) ([]Domain.Campaign, error) {

	query := `
	SELECT campaign_id, title, content, location, start_date, end_date, created_at
	FROM campaigns
	WHERE location LIKE CONCAT('%', $1, '%')
	AND end_date >= NOW()
	`

	rows, err := r.DB.Query(query, location)
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
		)

		if err != nil {
			return nil, err
		}

		campaigns = append(campaigns, c)
	}

	return campaigns, nil
}