CREATE TABLE IF NOT EXISTS blood_requests (
    request_id VARCHAR(36) PRIMARY KEY,
    hospital_id VARCHAR(36) NOT NULL,

    blood_type VARCHAR(3) NOT NULL,
    quantity INT NOT NULL,

    urgency_level VARCHAR(50) DEFAULT 'MEDIUM',

    status VARCHAR(50) DEFAULT 'PENDING',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (hospital_id) REFERENCES hospitals(hospital_id)
);