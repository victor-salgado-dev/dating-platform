-- 000007_create_conversations_and_messages.up.sql
--
-- Mensajería 1:1 (sin grupos en V1). Una conversación es un par de
-- usuarios ORDENADO CANÓNICAMENTE (user_one_id < user_two_id): así se
-- puede tener un UNIQUE simple en vez de tener que buscar con
-- "(a,b) OR (b,a)" en cada consulta. El código de aplicación siempre
-- ordena el par antes de insertar o buscar.

CREATE TABLE conversations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_one_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_two_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT conversations_distinct_users CHECK (user_one_id <> user_two_id),
    CONSTRAINT conversations_ordered_pair CHECK (user_one_id < user_two_id),
    UNIQUE (user_one_id, user_two_id)
);

CREATE INDEX conversations_user_one_id_idx ON conversations (user_one_id);
CREATE INDEX conversations_user_two_id_idx ON conversations (user_two_id);

CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body            TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    read_at         TIMESTAMPTZ,

    CONSTRAINT messages_body_length_check CHECK (char_length(body) > 0 AND char_length(body) <= 2000)
);

-- Paginación de una conversación, y localizar el último mensaje.
CREATE INDEX messages_conversation_id_created_at_idx ON messages (conversation_id, created_at);
