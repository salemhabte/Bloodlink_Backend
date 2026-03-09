CREATE TABLE IF NOT EXISTS notifications (
    notification_id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,

    type ENUM(
        'NORMAL',
        'EMERGENCY',
        'CONTRACT',
        'BLOOD_REQUEST'
    ) NOT NULL,

    title VARCHAR(255),
    message TEXT,

    is_read BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(user_id)
);