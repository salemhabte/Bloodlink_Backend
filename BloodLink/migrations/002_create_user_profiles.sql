CREATE TABLE IF NOT EXISTS User_Profile (
    profile_id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,

    full_name VARCHAR(255),
    phone VARCHAR(50),
    Address VARCHAR(100),
    profile_picture_url TEXT,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES Users(user_id)
);