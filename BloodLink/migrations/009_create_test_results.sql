CREATE TABLE IF NOT EXISTS donor_test_results (
    test_id VARCHAR(36) PRIMARY KEY,

    donation_id VARCHAR(36) NOT NULL,
    donor_id VARCHAR(36) NOT NULL,
    tested_by VARCHAR(36) NOT NULL,

    hiv_result BOOLEAN,
    hepatitis_result BOOLEAN,
    syphilis_result BOOLEAN,

    overall_status ENUM('SAFE','UNSAFE'),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (donation_id) REFERENCES donation_records(donation_id),
    FOREIGN KEY (donor_id) REFERENCES donors(donor_id),
    FOREIGN KEY (tested_by) REFERENCES users(user_id)
);