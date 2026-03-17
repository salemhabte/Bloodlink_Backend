CREATE TABLE donation_records (

    donation_id VARCHAR(36) PRIMARY KEY,

    donor_id VARCHAR(36) NOT NULL,

    -- blood collector user id
    collected_by VARCHAR(36) NOT NULL,

    collection_date DATE,

    -- screening results
    weight DECIMAL(5,2),
    blood_pressure VARCHAR(20),
    hemoglobin DECIMAL(4,2),
    temperature DECIMAL(4,2),
    pulse INT,

    quantity_ml INT,

    status ENUM(
        'PENDING',
        'APPROVED',
        'REJECTED_TEMPORARY',
        'REJECTED_PERMANENT'
    ),

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (donor_id) REFERENCES donors(donor_id),
    FOREIGN KEY (collected_by) REFERENCES users(user_id)
);