CREATE TABLE IF NOT EXISTS donors (
    donor_id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'Pending',  
    -- donor medical status:
-- Pending = newly registered / not yet medically reviewed
-- CLEARED = medically accepted
-- TEMPORARILY_DEFERRED = temporary restriction
-- PERMANENTLY_DEFERRED = permanent restriction

    blood_type VARCHAR(3) NULL,
    date_of_birth DATE NULL,
    weight DECIMAL(5,2) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(user_id)
);