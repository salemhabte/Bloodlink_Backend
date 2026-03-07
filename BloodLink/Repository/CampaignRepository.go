package Repository

import (
    "bloodlink/Domain"
    "database/sql"
    "time"

    "github.com/google/uuid"
)

// CampaignRepository implements Domain.ICampaignRepository
type CampaignRepository struct {
    DB *sql.DB
}

// NewCampaignRepository creates a new repository instance
func NewCampaignRepository(db *sql.DB) Domain.ICampaignRepository {
    return &CampaignRepository{DB: db}
}

func (r *CampaignRepository) CreateCampaign(campaign *Domain.Campaign) error {
    campaign.CampaignID = uuid.New().String()
    campaign.CreatedAt = time.Now()

    query := "INSERT INTO campaigns (campaign_id, title, content, location, start_date, end_date, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
    _, err := r.DB.Exec(query, campaign.CampaignID, campaign.Title, campaign.Content, campaign.Location, campaign.StartDate, campaign.EndDate, campaign.CreatedAt)
    return err
}

func (r *CampaignRepository) GetAllCampaigns() ([]Domain.Campaign, error) {
    query := "SELECT campaign_id, title, content, location, start_date, end_date, created_at FROM campaigns"
    rows, err := r.DB.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var campaigns []Domain.Campaign
    for rows.Next() {
        var c Domain.Campaign
        var startDateStr, endDateStr, createdAtStr string

        if err := rows.Scan(&c.CampaignID, &c.Title, &c.Content, &c.Location, &startDateStr, &endDateStr, &createdAtStr); err != nil {
            return nil, err
        }

        // Parse datetime strings to time.Time
        c.StartDate, err = time.Parse("2006-01-02 15:04:05", startDateStr)
        if err != nil {
            return nil, err
        }
        c.EndDate, err = time.Parse("2006-01-02 15:04:05", endDateStr)
        if err != nil {
            return nil, err
        }
        c.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
        if err != nil {
            return nil, err
        }

        campaigns = append(campaigns, c)
    }
    return campaigns, nil
}

func (r *CampaignRepository) GetCampaignByID(id string) (*Domain.Campaign, error) {
    query := "SELECT campaign_id, title, content, location, start_date, end_date, created_at FROM campaigns WHERE campaign_id=?"
    row := r.DB.QueryRow(query, id)

    var c Domain.Campaign
    var startDateStr, endDateStr, createdAtStr string

    if err := row.Scan(&c.CampaignID, &c.Title, &c.Content, &c.Location, &startDateStr, &endDateStr, &createdAtStr); err != nil {
        return nil, err
    }

    var err error
    c.StartDate, err = time.Parse("2006-01-02 15:04:05", startDateStr)
    if err != nil {
        return nil, err
    }
    c.EndDate, err = time.Parse("2006-01-02 15:04:05", endDateStr)
    if err != nil {
        return nil, err
    }
    c.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
    if err != nil {
        return nil, err
    }

    return &c, nil
}

func (r *CampaignRepository) UpdateCampaign(campaign *Domain.Campaign) error {
    // Fetch the existing campaign
    existing, err := r.GetCampaignByID(campaign.CampaignID)
    if err != nil {
        return err
    }

    // Update only the fields that are provided
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

    // Execute the update query with safe values
    query := `
        UPDATE campaigns 
        SET title=?, content=?, location=?, start_date=?, end_date=? 
        WHERE campaign_id=?
    `
    _, err = r.DB.Exec(query,
        existing.Title,
        existing.Content,
        existing.Location,
        existing.StartDate,
        existing.EndDate,
        existing.CampaignID,
    )
    return err
}

func (r *CampaignRepository) DeleteCampaign(id string) error {
    query := "DELETE FROM campaigns WHERE campaign_id=?"
    _, err := r.DB.Exec(query, id)
    return err
}