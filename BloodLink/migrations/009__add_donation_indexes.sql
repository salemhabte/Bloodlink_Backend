CREATE INDEX idx_donation_campaign_id
ON donation_records(campaign_id);

CREATE INDEX idx_donation_donor_id
ON donation_records(donor_id);

CREATE INDEX idx_donation_status
ON donation_records(status);

CREATE INDEX idx_donation_collection_date
ON donation_records(collection_date);