CREATE TABLE audit_logs (
    log_id SERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL, -- Admin/Staff who performed the action
    action VARCHAR(100) NOT NULL, -- e.g. "UPDATE_DONOR_STATUS", "DELETE_USER"
    target_type VARCHAR(50) NOT NULL, -- e.g. "DONOR", "DONATION", "USER"
    target_id VARCHAR(36) NOT NULL, -- UUID of the affected resource
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_target ON audit_logs(target_type, target_id);
