package Domain

import "time"

// Campaign represents a blood donation campaign
type Campaign struct {
    CampaignID string
    Title      string
    Content    string
    Location   string
    StartDate  time.Time
    EndDate    time.Time
    CreatedAt  time.Time
    IsDeleted  bool
}

type CampaignFilter struct {
    Title     string `json:"title"`
    Location  string `json:"location"`
    StartDate string `json:"start_date"`
    EndDate   string `json:"end_date"`
    LiveOnly  bool   `json:"live_only"`
}

type CampaignListResponse struct {
    TotalCampaigns       int        `json:"total_campaigns,omitempty"`
    OngoingCampaigns     int        `json:"ongoing_campaigns"`
    UpcomingCampaigns    int        `json:"upcoming_campaigns"`
    ClosingSoonCampaigns int        `json:"closing_soon_campaigns"`
    ClosedCampaigns      int        `json:"closed_campaigns"`
    Campaigns            []Campaign `json:"campaigns"`
}