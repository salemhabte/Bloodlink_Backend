-- Migration 039: Add license_document column to hospitals table

ALTER TABLE hospitals ADD COLUMN IF NOT EXISTS license_document VARCHAR(255);
