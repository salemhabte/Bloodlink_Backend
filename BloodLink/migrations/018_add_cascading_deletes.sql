-- Migration to add ON DELETE CASCADE safely using independent ALTER statements.
-- Idempotency is handled by the migration runner skipping "already exists" errors.

SET FOREIGN_KEY_CHECKS = 0;

-- 1. user_profiles
ALTER TABLE user_profiles DROP FOREIGN KEY user_profiles_ibfk_1;
ALTER TABLE user_profiles ADD CONSTRAINT fk_user_profile_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- 2. donors
ALTER TABLE donors DROP FOREIGN KEY donors_ibfk_1;
ALTER TABLE donors ADD CONSTRAINT fk_donor_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- 3. hospital_admins
ALTER TABLE hospital_admins DROP FOREIGN KEY hospital_admins_ibfk_1;
ALTER TABLE hospital_admins DROP FOREIGN KEY hospital_admins_ibfk_2;
ALTER TABLE hospital_admins ADD CONSTRAINT fk_hospital_admin_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
ALTER TABLE hospital_admins ADD CONSTRAINT fk_hospital_admin_hospital FOREIGN KEY (hospital_id) REFERENCES hospitals(hospital_id) ON DELETE CASCADE;

-- 4. blood_collectors
ALTER TABLE blood_collectors DROP FOREIGN KEY blood_collectors_ibfk_1;
ALTER TABLE blood_collectors DROP FOREIGN KEY blood_collectors_ibfk_2;
ALTER TABLE blood_collectors ADD CONSTRAINT fk_blood_collector_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
ALTER TABLE blood_collectors ADD CONSTRAINT fk_blood_collector_admin FOREIGN KEY (blood_bank_admin_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- 5. lab_technicians
ALTER TABLE lab_technicians DROP FOREIGN KEY lab_technicians_ibfk_1;
ALTER TABLE lab_technicians DROP FOREIGN KEY lab_technicians_ibfk_2;
ALTER TABLE lab_technicians ADD CONSTRAINT fk_lab_tech_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
ALTER TABLE lab_technicians ADD CONSTRAINT fk_lab_tech_admin FOREIGN KEY (blood_bank_admin_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- 6. donation_records
ALTER TABLE donation_records DROP FOREIGN KEY donation_records_ibfk_1;
ALTER TABLE donation_records DROP FOREIGN KEY donation_records_ibfk_2;
ALTER TABLE donation_records DROP FOREIGN KEY fk_collected_by_user;
ALTER TABLE donation_records ADD CONSTRAINT fk_donation_donor FOREIGN KEY (donor_id) REFERENCES donors(donor_id) ON DELETE CASCADE;
ALTER TABLE donation_records ADD CONSTRAINT fk_donation_collector FOREIGN KEY (collected_by) REFERENCES blood_collectors(collector_id) ON DELETE CASCADE;

-- 7. donor_test_results
ALTER TABLE donor_test_results DROP FOREIGN KEY donor_test_results_ibfk_1;
ALTER TABLE donor_test_results DROP FOREIGN KEY donor_test_results_ibfk_2;
ALTER TABLE donor_test_results DROP FOREIGN KEY donor_test_results_ibfk_3;
ALTER TABLE donor_test_results ADD CONSTRAINT fk_test_donation FOREIGN KEY (donation_id) REFERENCES donation_records(donation_id) ON DELETE CASCADE;
ALTER TABLE donor_test_results ADD CONSTRAINT fk_test_donor FOREIGN KEY (donor_id) REFERENCES donors(donor_id) ON DELETE CASCADE;
ALTER TABLE donor_test_results ADD CONSTRAINT fk_test_lab_tech FOREIGN KEY (tested_by) REFERENCES lab_technicians(lab_tech_id) ON DELETE CASCADE;

-- 8. blood_units
ALTER TABLE blood_units DROP FOREIGN KEY blood_units_ibfk_1;
ALTER TABLE blood_units ADD CONSTRAINT fk_blood_unit_donation FOREIGN KEY (donation_id) REFERENCES donation_records(donation_id) ON DELETE CASCADE;

-- 9. hospital_contracts
ALTER TABLE hospital_contracts DROP FOREIGN KEY hospital_contracts_ibfk_1;
ALTER TABLE hospital_contracts DROP FOREIGN KEY hospital_contracts_ibfk_2;
ALTER TABLE hospital_contracts ADD CONSTRAINT fk_contract_hospital FOREIGN KEY (hospital_id) REFERENCES hospitals(hospital_id) ON DELETE CASCADE;
ALTER TABLE hospital_contracts ADD CONSTRAINT fk_contract_admin FOREIGN KEY (blood_bank_admin_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- 10. emergency_requests
ALTER TABLE emergency_requests DROP FOREIGN KEY emergency_requests_ibfk_1;
ALTER TABLE emergency_requests ADD CONSTRAINT fk_emergency_admin FOREIGN KEY (blood_bank_admin_id) REFERENCES users(user_id) ON DELETE CASCADE;

-- 11. notifications
ALTER TABLE notifications DROP FOREIGN KEY notifications_ibfk_1;
ALTER TABLE notifications ADD CONSTRAINT fk_notification_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;

SET FOREIGN_KEY_CHECKS = 1;
