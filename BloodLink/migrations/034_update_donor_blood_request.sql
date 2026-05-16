-- 034_update_donor_blood_request.sql

-- 1. Rename quantity_ml to units
ALTER TABLE donor_blood_requests RENAME COLUMN quantity_ml TO units;

-- 2. Add component_type
ALTER TABLE donor_blood_requests ADD COLUMN component_type VARCHAR(50);

-- Since existing records might not have a component type but are historical, we can leave it NULL for old ones,
-- or set a default like 'WHOLE_BLOOD'. Leaving as NULL is safer for historical accuracy unless NOT NULL is required.
