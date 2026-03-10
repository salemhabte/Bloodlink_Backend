CREATE TABLE IF NOT EXISTS Donors (
    donor_id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,

    blood_type VARCHAR(3) NOT NULL,
    date_of_birth DATE NOT NULL,
    weight DECIMAL(5,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);