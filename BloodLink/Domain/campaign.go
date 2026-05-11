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
}

type CampaignFilter struct {
    Location  string `json:"location"`
    StartDate string `json:"start_date"`
    EndDate   string `json:"end_date"`
}