-- 000010_create_consents.up.sql
--
-- Registro de consentimientos (sección 14): fecha/hora y versión del
-- documento aceptado, separado por tipo de documento a propósito —
-- aceptar los Términos no debe asumirse como aceptación automática de
-- cualquier otro tratamiento de datos. Es un registro de auditoría:
-- solo se inserta, nunca se actualiza ni se borra (salvo cascada por
-- eliminación de la cuenta).

CREATE TABLE consents (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    document_type     TEXT NOT NULL,
    document_version  TEXT NOT NULL,
    accepted_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT consents_document_type_check CHECK (document_type IN ('terms', 'privacy_policy'))
);

CREATE INDEX consents_user_id_idx ON consents (user_id);
