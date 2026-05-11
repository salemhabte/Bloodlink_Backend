-- Expand status column to hold longer status values like 'PARTIALLY APPROVED', 'PARTIALLY FULFILLED'
ALTER TABLE donor_blood_requests
    ALTER COLUMN status TYPE VARCHAR(50);

-- Add index for faster admin filtering by status and blood_type
CREATE INDEX IF NOT EXISTS idx_donor_blood_requests_status
    ON donor_blood_requests (status);

CREATE INDEX IF NOT EXISTS idx_donor_blood_requests_blood_type
    ON donor_blood_requests (blood_type);

CREATE INDEX IF NOT EXISTS idx_donor_blood_requests_created_at
    ON donor_blood_requests (created_at);
