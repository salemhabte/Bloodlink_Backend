-- 035_add_reserved_units_to_donor_requests.sql
ALTER TABLE donor_blood_requests ADD COLUMN IF NOT EXISTS reserved_units INTEGER NOT NULL DEFAULT 0;
