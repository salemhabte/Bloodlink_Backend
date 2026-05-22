-- Create sequences tracking table
CREATE TABLE id_counters (
    counter_type VARCHAR(50) NOT NULL, -- 'DONATION' or 'BLOOD_UNIT'
    counter_year INT NOT NULL,
    current_value INT NOT NULL DEFAULT 0,
    PRIMARY KEY (counter_type, counter_year)
);

-- Add donation_number to donation_records
ALTER TABLE donation_records 
ADD COLUMN donation_number VARCHAR(50);

ALTER TABLE donation_records 
ADD CONSTRAINT unique_donation_number UNIQUE (donation_number);

-- Add unit_number to blood_units
ALTER TABLE blood_units 
ADD COLUMN unit_number VARCHAR(50);

ALTER TABLE blood_units 
ADD CONSTRAINT unique_unit_number UNIQUE (unit_number);
