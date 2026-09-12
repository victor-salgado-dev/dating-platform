-- 000009_add_user_role.up.sql
--
-- Añade el rol de cuenta, necesario para el panel de administración
-- (Fase 10). Cambio aditivo con DEFAULT: no rompe filas existentes.

ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';

ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('user', 'admin'));

-- Índice parcial: solo hay unos pocos admins, no hace falta indexar 'user'.
CREATE INDEX users_role_admin_idx ON users (role) WHERE role = 'admin';
