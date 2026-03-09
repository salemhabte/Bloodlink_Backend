CREATE TABLE IF NOT EXISTS emergency_requests (
    emergency_id VARCHAR(36) PRIMARY KEY,

    blood_bank_admin_id VARCHAR(36) NOT NULL,

    blood_type VARCHAR(3) NOT NULL,
    hospital_name VARCHAR(255),
    hospital_location VARCHAR(255),

    required_units INT,

    urgency_level ENUM('LOW','MEDIUM','HIGH') DEFAULT 'HIGH',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (blood_bank_admin_id) REFERENCES users(user_id)
);