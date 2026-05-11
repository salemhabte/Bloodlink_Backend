CREATE TABLE donor_blood_requests (
    request_id VARCHAR(36) PRIMARY KEY,
    donor_id VARCHAR(36) NOT NULL,
    blood_type VARCHAR(3) NOT NULL,
    quantity_ml INT NOT NULL,
    reason TEXT,
    priority_score INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'PENDING',  /* PENDING
                                              PARTIALLY_APPROVED
                                              APPROVED
                                              PARTIALLY_FULFILLED
                                              FULFILLED
                                              REJECTED */
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (donor_id) REFERENCES donors(donor_id)
);  