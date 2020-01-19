ALTER TABLE billing_plans ADD COLUMN upgrade_rate integer;

UPDATE billing_plans SET upgrade_rate = id;