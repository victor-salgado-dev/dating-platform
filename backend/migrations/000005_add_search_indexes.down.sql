-- 000005_add_search_indexes.down.sql

DROP INDEX IF EXISTS profiles_interests_gin_idx;
DROP INDEX IF EXISTS profiles_languages_gin_idx;
DROP INDEX IF EXISTS profiles_wants_children_idx;
DROP INDEX IF EXISTS profiles_has_children_idx;
DROP INDEX IF EXISTS profiles_relationship_goal_idx;
