-- Add position_number to blood_units table
ALTER TABLE blood_units
  ADD COLUMN IF NOT EXISTS position_number VARCHAR(50);
