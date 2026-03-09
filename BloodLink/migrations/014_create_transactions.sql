CREATE TABLE IF NOT EXISTS transactions (
    transaction_id VARCHAR(36) PRIMARY KEY,
    request_id VARCHAR(36) NOT NULL,
    dispatch_date TIMESTAMP NULL,
    status ENUM('PENDING','DISPATCHED','COMPLETED') DEFAULT 'PENDING',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
    FOREIGN KEY (request_id) REFERENCES blood_requests(request_id)
);