-- 000008_create_blocks_and_reports.up.sql

-- Bloqueo: relación dirigida (importa quién bloqueó a quién, para poder
-- desbloquear solo desde el lado que bloqueó y listar "a quién he
-- bloqueado yo"). El EFECTO de visibilidad, en cambio, es mutuo: los
-- módulos que consultan esta tabla (search, favorites, messaging,
-- profiles) comprueban el bloqueo en AMBOS sentidos.
CREATE TABLE blocks (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT blocks_distinct_users CHECK (blocker_id <> blocked_id),
    UNIQUE (blocker_id, blocked_id)
);

CREATE INDEX blocks_blocker_id_idx ON blocks (blocker_id);
-- Acelera las comprobaciones "¿me ha bloqueado esta persona?" en el
-- sentido inverso, usadas por search/favorites/messaging/profiles.
CREATE INDEX blocks_blocked_id_idx ON blocks (blocked_id);

-- Reportes: solo se crean en esta fase. La revisión/gestión (panel de
-- moderación) es la Fase 10; por eso ya incluye `status`, pero nada
-- todavía lo cambia de 'pending'.
CREATE TABLE reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reported_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason      TEXT NOT NULL,
    description TEXT,
    status      TEXT NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT reports_distinct_users CHECK (reporter_id <> reported_id),
    CONSTRAINT reports_reason_check CHECK (reason IN (
        'spam', 'fake_profile', 'harassment', 'inappropriate_content', 'underage', 'other'
    )),
    CONSTRAINT reports_status_check CHECK (status IN ('pending', 'reviewed', 'dismissed')),
    CONSTRAINT reports_description_length_check CHECK (description IS NULL OR char_length(description) <= 2000)
);

CREATE INDEX reports_status_idx ON reports (status);
CREATE INDEX reports_reported_id_idx ON reports (reported_id);
