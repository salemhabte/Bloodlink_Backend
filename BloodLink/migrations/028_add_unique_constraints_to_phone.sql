-- Add unique constraints to phone numbers

-- For users table
ALTER TABLE users
ADD CONSTRAINT uni_users_phone UNIQUE (phone);

-- For hospitals table
ALTER TABLE hospitals
ADD CONSTRAINT uni_hospitals_phone UNIQUE (phone);

-- For user_profiles table (if it's still being used for phone)
ALTER TABLE user_profiles
ADD CONSTRAINT uni_user_profiles_phone UNIQUE (phone);
