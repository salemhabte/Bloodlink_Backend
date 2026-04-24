CREATE TYPE blood_type_enum AS ENUM ('A+','A-','B+','B-','AB+','AB-','O+','O-');
CREATE TYPE urgency_level_enum AS ENUM ('LOW','MEDIUM','HIGH','CRITICAL');
CREATE TYPE blood_request_status_enum AS ENUM ('PENDING', 'APPROVED_PARTIALLY_FULFILLED', 'REJECTED', 'FULFILLED');

ALTER TABLE blood_requests
    ALTER COLUMN blood_type TYPE blood_type_enum USING blood_type::text::blood_type_enum,
    ALTER COLUMN urgency_level DROP DEFAULT,
    ALTER COLUMN urgency_level TYPE urgency_level_enum USING urgency_level::text::urgency_level_enum,
    ALTER COLUMN urgency_level SET DEFAULT 'MEDIUM',
    ALTER COLUMN status DROP DEFAULT,
    ALTER COLUMN status TYPE blood_request_status_enum USING status::text::blood_request_status_enum,
    ALTER COLUMN status SET DEFAULT 'PENDING',
    ADD COLUMN IF NOT EXISTS approved_at TIMESTAMP NULL;
