CREATE TABLE IF NOT EXISTS hospital_requests (
    request_id VARCHAR(36) PRIMARY KEY,
    hospital_name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    phone VARCHAR(50) NOT NULL,
    license_document VARCHAR(255),
    status VARCHAR(50) DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS hospital_request_admins (
    request_admin_id VARCHAR(36) PRIMARY KEY,
    request_id VARCHAR(36) NOT NULL,
    admin_full_name VARCHAR(255) NOT NULL,
    admin_email VARCHAR(255) NOT NULL UNIQUE,
    admin_phone VARCHAR(50) NOT NULL,
    admin_password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (request_id) REFERENCES hospital_requests(request_id) ON DELETE CASCADE
);
