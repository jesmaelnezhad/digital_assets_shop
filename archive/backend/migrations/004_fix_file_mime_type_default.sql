-- Fix: set default empty string for file_mime_type and update existing NULLs
ALTER TABLE products ALTER COLUMN file_mime_type SET DEFAULT '';
UPDATE products SET file_mime_type = '' WHERE file_mime_type IS NULL;
