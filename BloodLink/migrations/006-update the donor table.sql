ALTER TABLE donors RENAME COLUMN status TO overall_status;
ALTER TABLE donors ALTER COLUMN overall_status TYPE VARCHAR(20);
ALTER TABLE donors ALTER COLUMN overall_status SET DEFAULT 'Pending';