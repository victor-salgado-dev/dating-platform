-- 000001_initial_schema.up.sql
--
-- Migración inicial. Prepara PostgreSQL con las extensiones básicas
-- que usará el resto del esquema (generación de UUIDs).
-- No crea todavía tablas de dominio: eso corresponde a la Fase 2.

CREATE EXTENSION IF NOT EXISTS pgcrypto;
