-- Add rejection_reason to donation_records
ALTER TABLE donation_records ADD COLUMN rejection_reason VARCHAR(255);

-- Add is_deleted to blood_units for soft deletes
ALTER TABLE blood_units ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;
