-- 000009_add_user_role.down.sql

DROP INDEX IF EXISTS users_role_admin_idx;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users DROP COLUMN IF EXISTS role;
