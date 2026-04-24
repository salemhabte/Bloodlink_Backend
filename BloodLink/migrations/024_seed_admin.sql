INSERT INTO users (user_id, email, full_name, phone, password_hash, role, is_active)
VALUES ('00000000-0000-0000-0000-000000000000', 'admin@bloodlink.com', 'System Admin', '+1234567890', 'none', 'bloodbankadmin', true)
ON CONFLICT (user_id) DO NOTHING;
