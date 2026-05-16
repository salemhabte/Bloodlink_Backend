-- Add is_deleted to campaigns for soft deletes
ALTER TABLE campaigns ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE;
