-- =========================================
-- INDEXES FOR donation_records
-- =========================================

CREATE INDEX IF NOT EXISTS idx_donation_records_collected_by
ON donation_records(collected_by);


-- =========================================
-- INDEXES FOR donors
-- =========================================

CREATE INDEX IF NOT EXISTS idx_donors_user_id
ON donors(user_id);

CREATE INDEX IF NOT EXISTS idx_donors_overall_status
ON donors(overall_status);

CREATE INDEX IF NOT EXISTS idx_donors_blood_type
ON donors(blood_type);


-- =========================================
-- INDEXES FOR donor_test_results
-- =========================================

CREATE INDEX IF NOT EXISTS idx_test_donation_id
ON donor_test_results(donation_id);

CREATE INDEX IF NOT EXISTS idx_test_donor_id
ON donor_test_results(donor_id);

CREATE INDEX IF NOT EXISTS idx_test_tested_by
ON donor_test_results(tested_by);

CREATE INDEX IF NOT EXISTS idx_test_overall_status
ON donor_test_results(overall_status);


-- =========================================
-- INDEXES FOR blood_units (FULL SET)
-- =========================================

CREATE INDEX IF NOT EXISTS idx_blood_units_donation_id
ON blood_units(donation_id);

CREATE INDEX IF NOT EXISTS idx_blood_units_blood_type
ON blood_units(blood_type);

CREATE INDEX IF NOT EXISTS idx_blood_units_status
ON blood_units(status);

CREATE INDEX IF NOT EXISTS idx_blood_units_expiration_date
ON blood_units(expiration_date);