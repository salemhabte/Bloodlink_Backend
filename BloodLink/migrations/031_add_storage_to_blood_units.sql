-- Add storage fields to blood_units
ALTER TABLE blood_units
  ADD COLUMN IF NOT EXISTS storage_location VARCHAR(100),
  ADD COLUMN IF NOT EXISTS rack_number      VARCHAR(50),
  ADD COLUMN IF NOT EXISTS shelf_number     VARCHAR(50);
