-- Migration 040: Clean up partially created records from previous failed approval of request '24951f72-6d5b-465b-8f02-0b6b66278d35'

-- 1. Delete from hospital_admins first
DELETE FROM hospital_admins WHERE user_id IN (SELECT user_id FROM users WHERE email IN (SELECT admin_email FROM hospital_request_admins WHERE request_id = '24951f72-6d5b-465b-8f02-0b6b66278d35'));

-- 2. Delete from users
DELETE FROM users WHERE email IN (SELECT admin_email FROM hospital_request_admins WHERE request_id = '24951f72-6d5b-465b-8f02-0b6b66278d35');

-- 3. Delete from hospitals
DELETE FROM hospitals WHERE phone IN (SELECT phone FROM hospital_requests WHERE request_id = '24951f72-6d5b-465b-8f02-0b6b66278d35');
