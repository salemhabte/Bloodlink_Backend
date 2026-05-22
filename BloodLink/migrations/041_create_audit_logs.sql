CREATE TABLE IF NOT EXISTS admin_audit_logs (
    log_id VARCHAR(36) PRIMARY KEY,
    admin_id VARCHAR(36) REFERENCES users(user_id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(100) NOT NULL,
    target_id VARCHAR(36),
    details TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
