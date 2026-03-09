CREATE TABLE IF NOT EXISTS blood_collectors (
    collector_id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    blood_bank_admin_id VARCHAR(36) NOT NULL,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (blood_bank_admin_id) REFERENCES users(user_id)
);