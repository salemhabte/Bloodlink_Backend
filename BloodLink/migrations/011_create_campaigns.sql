CREATE TABLE IF NOT EXISTS campaigns (
    campaign_id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255),
    content TEXT,
    location VARCHAR(255),
    start_date DATETIME,
    end_date DATETIME,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);