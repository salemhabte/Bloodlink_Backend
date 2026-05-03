ALTER TABLE hospital_contracts
ADD COLUMN IF NOT EXISTS hospital_signature_path VARCHAR(255),
ADD COLUMN IF NOT EXISTS admin_signature_path VARCHAR(255);
