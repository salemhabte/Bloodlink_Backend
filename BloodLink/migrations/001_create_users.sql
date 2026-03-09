CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,

    role ENUM(
        'DONOR',
        'BLOODBANK_ADMIN',
        'BLOOD_COLLECTOR',
        'LAB_TECHNICIAN',
        'HOSPITAL_ADMIN'
    ) NOT NULL,

    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);