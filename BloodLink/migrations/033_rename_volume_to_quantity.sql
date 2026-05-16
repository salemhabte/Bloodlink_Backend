ALTER TABLE blood_units RENAME COLUMN volume_ml TO quantity_ml;
ALTER TABLE inventory_audit RENAME COLUMN volume_ml TO quantity_ml;
ALTER TABLE blood_requests RENAME COLUMN fulfilled_volume_ml TO fulfilled_quantity_ml;
