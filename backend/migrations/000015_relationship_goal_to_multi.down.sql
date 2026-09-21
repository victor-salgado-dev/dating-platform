-- 000015_relationship_goal_to_multi.down.sql
--
-- Nota: si alguien llegó a marcar más de un objetivo, al volver atrás
-- solo se conserva el primero del array — es la única forma sensata de
-- volver a un solo valor sin inventarse cuál "vale más".

ALTER TABLE profiles ADD COLUMN relationship_goal TEXT
    CONSTRAINT profiles_relationship_goal_check
    CHECK (relationship_goal IS NULL OR relationship_goal IN
        ('casual', 'long_term', 'friendship', 'marriage', 'not_sure'));

UPDATE profiles SET relationship_goal = relationship_goals[1]
    WHERE relationship_goals IS NOT NULL AND array_length(relationship_goals, 1) > 0;

ALTER TABLE profiles DROP COLUMN relationship_goals;
