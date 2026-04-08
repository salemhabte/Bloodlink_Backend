CREATE TABLE IF NOT EXISTS blood_units (
    blood_unit_id VARCHAR(36) PRIMARY KEY,

    donation_id VARCHAR(36) NOT NULL,

    blood_type VARCHAR(3),
    volume_ml INT,

    collection_date DATE,
    expiration_date DATE,

    status VARCHAR(50) DEFAULT 'AVAILABLE',

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (donation_id) REFERENCES donation_records(donation_id)
);