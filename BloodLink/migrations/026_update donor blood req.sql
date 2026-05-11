ALTER TABLE donor_blood_requests

ADD COLUMN donor_name VARCHAR(255),
ADD COLUMN donor_email VARCHAR(255),
ADD COLUMN donor_phone VARCHAR(50),
ADD COLUMN donor_address VARCHAR(255),

ADD COLUMN hospital_name VARCHAR(255),
ADD COLUMN hospital_address VARCHAR(255),
ADD COLUMN hospital_phone VARCHAR(50);