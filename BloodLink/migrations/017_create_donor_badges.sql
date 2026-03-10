CREATE TABLE IF NOT EXISTS donor_badges (
    badge_id VARCHAR(36) PRIMARY KEY,
    donor_id VARCHAR(36) NOT NULL,
    badge_name VARCHAR(255) NOT NULL,
    description TEXT,
    awarded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_donor_badge FOREIGN KEY (donor_id) REFERENCES donors(donor_id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);