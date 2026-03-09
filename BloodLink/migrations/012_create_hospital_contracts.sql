CREATE TABLE IF NOT EXISTS hospital_contracts (
    contract_id VARCHAR(36) PRIMARY KEY,
    hospital_id VARCHAR(36) NOT NULL,
    blood_bank_admin_id VARCHAR(36) NOT NULL,

    document VARCHAR(255),

    status ENUM(
        'PENDING',
        'APPROVED',
        'REJECTED'
    ) DEFAULT 'PENDING',

    contract_start DATE,
    contract_end DATE,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (hospital_id) REFERENCES hospitals(hospital_id),
    FOREIGN KEY (blood_bank_admin_id) REFERENCES users(user_id)
);