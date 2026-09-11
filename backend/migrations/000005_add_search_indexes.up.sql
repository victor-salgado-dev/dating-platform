-- 000005_add_search_indexes.up.sql
--
-- Índices pensados directamente para los filtros de búsqueda (Fase 5).
-- Los índices parciales (WHERE ... IS NOT NULL) evitan indexar las filas
-- que de todos modos quedan excluidas por la regla de los datos
-- faltantes cuando se filtra por ese campo.

CREATE INDEX profiles_relationship_goal_idx
    ON profiles (relationship_goal)
    WHERE relationship_goal IS NOT NULL;

CREATE INDEX profiles_has_children_idx
    ON profiles (has_children)
    WHERE has_children IS NOT NULL;

CREATE INDEX profiles_wants_children_idx
    ON profiles (wants_children)
    WHERE wants_children IS NOT NULL;

-- GIN para los operadores de solape (&&) usados en los filtros de
-- idiomas e intereses.
CREATE INDEX profiles_languages_gin_idx ON profiles USING GIN (languages);
CREATE INDEX profiles_interests_gin_idx ON profiles USING GIN (interests);
