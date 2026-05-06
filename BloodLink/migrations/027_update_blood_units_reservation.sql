-- Add reservation tracking to blood_units
ALTER TABLE blood_units
    ADD COLUMN IF NOT EXISTS reserved_for_hospital_id VARCHAR(36) NULL,
    ADD COLUMN IF NOT EXISTS reserved_at TIMESTAMP NULL,
    ADD COLUMN IF NOT EXISTS request_id VARCHAR(36) NULL;

-- Add fulfillment details to blood_requests
ALTER TABLE blood_requests
    ADD COLUMN IF NOT EXISTS fulfilled_volume_ml INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fulfilled_count INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS notes TEXT NULL;

-- Audit log for deleted blood units (preserves analytics after deletion)
CREATE TABLE IF NOT EXISTS inventory_audit (
    id SERIAL PRIMARY KEY,
    blood_unit_id   VARCHAR(36)  NOT NULL,
    blood_type      VARCHAR(10),
    volume_ml       INT,
    status_at_deletion VARCHAR(50),
    deleted_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
