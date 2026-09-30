-- 000020_perf_indexes_and_popularity.up.sql
--
-- 1. Índices que faltaban para las pestañas de Home:
--    - "new members" ordena por profiles.created_at (el índice de la 000019
--      está sobre users.created_at y esa consulta no lo usa).
--    - "online now" filtra/ordena por users.last_active_at.
--    - El filtro de relationship_goals usa && sobre un text[]; los GIN
--      antiguos desaparecieron con las columnas (000014, 000015, 000016).
CREATE INDEX IF NOT EXISTS profiles_created_at_idx
    ON profiles (created_at DESC, id);

CREATE INDEX IF NOT EXISTS users_last_active_at_idx
    ON users (last_active_at DESC)
    WHERE last_active_at IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS profiles_relationship_goals_gin_idx
    ON profiles USING GIN (relationship_goals);

-- 2. Escribir last_active_at no debe cambiar updated_at (se toca cada pocos
--    minutos por usuario activo). El trigger se recrea con una condición.
DROP TRIGGER IF EXISTS users_set_updated_at ON users;
CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    WHEN (OLD.last_active_at IS NOT DISTINCT FROM NEW.last_active_at)
    EXECUTE FUNCTION set_updated_at();

-- 3. Popularidad: antes se calculaba con 4 subconsultas correlacionadas por
--    cada perfil en cada petición y contando todo el histórico. Ahora es una
--    vista materializada (actividad de los últimos 30 días) que el backend
--    refresca cada pocos minutos (search.StartPopularityRefresher).
CREATE MATERIALIZED VIEW profile_popularity AS
SELECT e.profile_id, COUNT(*)::int AS score
FROM (
    SELECT l.to_profile_id AS profile_id
      FROM likes l
     WHERE l.created_at >= now() - interval '30 days'
    UNION ALL
    SELECT f.favorite_profile_id
      FROM favorites f
     WHERE f.created_at >= now() - interval '30 days'
    UNION ALL
    SELECT v.visited_profile_id
      FROM profile_visits v
     WHERE v.visited_at >= now() - interval '30 days'
    UNION ALL
    SELECT p.id
      FROM messages m
      JOIN conversations c ON c.id = m.conversation_id
      JOIN profiles p ON p.user_id = CASE WHEN m.sender_id = c.user_one_id
                                          THEN c.user_two_id
                                          ELSE c.user_one_id END
     WHERE m.created_at >= now() - interval '30 days'
) e
GROUP BY e.profile_id;

-- Necesario para REFRESH ... CONCURRENTLY (no bloquea lecturas).
CREATE UNIQUE INDEX profile_popularity_profile_id_idx ON profile_popularity (profile_id);
CREATE INDEX profile_popularity_score_idx ON profile_popularity (score DESC);
