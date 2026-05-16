-- Rename status values for consistency
UPDATE donation_records SET status = 'REJECTED_TEMPORARY' WHERE status = 'TEMPORARILY_REJECTED';
UPDATE donors SET overall_status = 'TEMPORARILY_DEFERRED' WHERE overall_status = 'DEFERRED_TEMPORARY';
UPDATE donor_test_results SET overall_status = 'TEMPORARILY_DEFERRED' WHERE overall_status = 'DEFERRED_TEMPORARY';
