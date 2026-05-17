ALTER TABLE blood_requests
ADD COLUMN IF NOT EXISTS component VARCHAR(50) DEFAULT 'Whole Blood';

ALTER TYPE urgency_level_enum ADD VALUE IF NOT EXISTS 'emergency';
ALTER TYPE urgency_level_enum ADD VALUE IF NOT EXISTS 'normal';
