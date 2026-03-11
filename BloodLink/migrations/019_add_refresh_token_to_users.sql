-- Migration to add refresh_token for session management/logout revocation
ALTER TABLE users ADD COLUMN refresh_token TEXT;
