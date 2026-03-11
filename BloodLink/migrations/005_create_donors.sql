CREATE TABLE IF NOT EXISTS donors (
    donor_id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'Pending',

    blood_type VARCHAR(3) NULL,
    date_of_birth DATE NULL,
    weight DECIMAL(5,2) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(user_id)
);