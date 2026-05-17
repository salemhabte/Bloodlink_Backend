-- Migration 038: Add unique constraints to hospital requests tables for phone numbers

-- 1. Clean up duplicate admin_phones in hospital_request_admins (keeping the newest request, cascading delete)
DELETE FROM hospital_requests
WHERE request_id IN (
    SELECT request_id 
    FROM hospital_request_admins
    WHERE request_admin_id NOT IN (
        SELECT DISTINCT ON (admin_phone) request_admin_id 
        FROM hospital_request_admins 
        ORDER BY admin_phone, created_at DESC
    )
);

-- 2. Clean up duplicate hospital phones in hospital_requests where status is PENDING (keeping the newest request)
DELETE FROM hospital_requests 
WHERE status = 'PENDING' 
  AND request_id NOT IN (
    SELECT DISTINCT ON (phone) request_id 
    FROM hospital_requests 
    WHERE status = 'PENDING' 
    ORDER BY phone, created_at DESC
  );

-- Enforce unique constraint for pending phone numbers in hospital_requests
CREATE UNIQUE INDEX IF NOT EXISTS uni_pending_hospital_requests_phone ON hospital_requests (phone) WHERE status = 'PENDING';

-- Enforce unique constraint for admin_phone in hospital_request_admins
ALTER TABLE hospital_request_admins ADD CONSTRAINT uni_hospital_request_admins_phone UNIQUE (admin_phone);
