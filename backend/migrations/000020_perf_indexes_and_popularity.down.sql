DROP MATERIALIZED VIEW IF EXISTS profile_popularity;

DROP TRIGGER IF EXISTS users_set_updated_at ON users;
CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

DROP INDEX IF EXISTS profiles_relationship_goals_gin_idx;
DROP INDEX IF EXISTS users_last_active_at_idx;
DROP INDEX IF EXISTS profiles_created_at_idx;
