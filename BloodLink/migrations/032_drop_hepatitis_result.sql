-- Drop the old hepatitis_result column now that data has been migrated
ALTER TABLE donor_test_results DROP COLUMN IF EXISTS hepatitis_result;
