ALTER TABLE donation_records
ADD COLUMN campaign_id VARCHAR(36) NULL;

-- Add foreign key constraint
ALTER TABLE donation_records
ADD CONSTRAINT fk_campaign
FOREIGN KEY (campaign_id) REFERENCES campaigns(campaign_id)
ON DELETE SET NULL;

-- Add index
CREATE INDEX idx_donation_campaign_id
ON donation_records(campaign_id);
