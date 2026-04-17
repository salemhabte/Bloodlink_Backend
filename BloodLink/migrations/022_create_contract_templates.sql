CREATE TABLE IF NOT EXISTS contract_templates (
    template_id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed a sample template
INSERT INTO contract_templates (template_id, name, content, created_by, created_at)
VALUES (
    'system-default-template',
    'Standard Central Blood Supply Contract',
    'This blood supply contract is made and entered into on {{contract_start_date}} between the Centralized Blood Bank and {{hospital_name}}.\n\nThe contract binds the aforementioned hospital to the terms of safe blood transfer and regulatory compliance, valid until {{contract_end_date}}.',
    NULL,
    CURRENT_TIMESTAMP
) ON CONFLICT (template_id) DO NOTHING;

ALTER TABLE hospital_contracts
ADD COLUMN IF NOT EXISTS template_id VARCHAR(36);
