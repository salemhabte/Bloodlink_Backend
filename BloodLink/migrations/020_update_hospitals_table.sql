ALTER TABLE hospitals
CHANGE name hospital_name VARCHAR(255) NOT NULL;

ALTER TABLE hospitals
ADD COLUMN city VARCHAR(100),
ADD COLUMN contact_person_name VARCHAR(255),
ADD COLUMN contact_person_phone VARCHAR(50),
ADD COLUMN status VARCHAR(50) DEFAULT 'PENDING';
