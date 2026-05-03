DROP TABLE IF EXISTS emergency_requests;

CREATE TABLE emergency_requests (
    emergency_id VARCHAR(36) PRIMARY KEY,
    request_id VARCHAR(36), -- Optional, linked to a hospital request
    blood_type VARCHAR(3) NOT NULL,
    quantity_required INT NOT NULL,
    quantity_fulfilled INT DEFAULT 0,
    urgency_level VARCHAR(50),
    hospital_name VARCHAR(255),
    location VARCHAR(255),
    status VARCHAR(20) DEFAULT 'PENDING_PUBLISH',
    is_manual BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    
    FOREIGN KEY (request_id) REFERENCES blood_requests(request_id) ON DELETE SET NULL
);
