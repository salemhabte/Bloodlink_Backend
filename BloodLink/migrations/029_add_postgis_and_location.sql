-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Add location columns to user_profiles (for donors)
ALTER TABLE user_profiles 
ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS location_geo GEOGRAPHY(POINT, 4326);

-- Add location columns to hospitals
ALTER TABLE hospitals
ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS location_geo GEOGRAPHY(POINT, 4326);

-- Add location columns to hospital_requests
ALTER TABLE hospital_requests
ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS location_geo GEOGRAPHY(POINT, 4326);

-- Add location columns to emergency_requests
ALTER TABLE emergency_requests
ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS location_geo GEOGRAPHY(POINT, 4326);

-- Create spatial indexes
CREATE INDEX IF NOT EXISTS idx_user_profiles_location ON user_profiles USING GIST (location_geo);
CREATE INDEX IF NOT EXISTS idx_hospitals_location ON hospitals USING GIST (location_geo);
CREATE INDEX IF NOT EXISTS idx_emergency_requests_location ON emergency_requests USING GIST (location_geo);
