CREATE TABLE IF NOT EXISTS hospital_admins (
    hospital_admin_id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    hospital_id VARCHAR(36) NOT NULL,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (hospital_id) REFERENCES hospitals(hospital_id)
);