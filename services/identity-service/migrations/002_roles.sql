ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(32) NOT NULL DEFAULT 'customer';
ALTER TABLE users ADD COLUMN IF NOT EXISTS staff_tabs TEXT NOT NULL DEFAULT '';
UPDATE users SET role = 'customer' WHERE role IS NULL OR role = '' OR role = 'user';
UPDATE users SET role = 'admin', staff_tabs = '' WHERE lower(email) = 'nia@example.com';
UPDATE users SET role = 'staff', staff_tabs = 'Products,Banner,Orders,Community'
  WHERE lower(email) = 'leo@example.com' AND role <> 'admin';
