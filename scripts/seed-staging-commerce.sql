-- Apply to appdb_commerce_staging.

INSERT INTO coupons (code, discount_type, discount_value, min_purchase, max_uses, current_uses, active)
VALUES
('SAVE12', 'percentage', 12, 0, 0, 0, true),
('WELCOME', 'percentage', 10, 0, 0, 0, true),
('MARBLE', 'fixed', 5, 0, 0, 0, true)
ON CONFLICT (code) DO UPDATE SET discount_type = EXCLUDED.discount_type, discount_value = EXCLUDED.discount_value, active = true;
