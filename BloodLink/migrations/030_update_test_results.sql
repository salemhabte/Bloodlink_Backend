-- Add hepatitis_b and hepatitis_c columns, copy existing data, drop component_type
ALTER TABLE donor_test_results
  ADD COLUMN IF NOT EXISTS hepatitis_b_result VARCHAR(50),
  ADD COLUMN IF NOT EXISTS hepatitis_c_result VARCHAR(50);

-- Migrate existing hepatitis_result → hepatitis_b_result
UPDATE donor_test_results
SET hepatitis_b_result = hepatitis_result
WHERE hepatitis_b_result IS NULL AND hepatitis_result IS NOT NULL;

-- Drop component_type (belongs in blood_units, not test results)
ALTER TABLE donor_test_results DROP COLUMN IF EXISTS component_type;
