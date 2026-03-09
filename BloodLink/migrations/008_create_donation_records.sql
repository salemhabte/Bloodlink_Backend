CREATE TABLE IF NOT EXISTS donation_records (
    donation_id VARCHAR(36) PRIMARY KEY,

    donor_id VARCHAR(36) NOT NULL,
    collected_by VARCHAR(36) NOT NULL,

    collection_date DATE,
    quantity_ml INT,

    status ENUM(
        'PENDING',
        'APPROVED',
        'REJECTED_TEMPORARY',
        'REJECTED_PERMANENT'
    ),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (donor_id) REFERENCES donors(donor_id),
    FOREIGN KEY (collected_by) REFERENCES blood_collectors(collector_id)
);