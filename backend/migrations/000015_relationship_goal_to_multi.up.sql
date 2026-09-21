-- 000015_relationship_goal_to_multi.up.sql
--
-- "Lo que busco en una relación" pasa de single-select a multi-select:
-- alguien puede estar abierto a la vez a "casual" y a "long_term", por
-- ejemplo. Mismo patrón que body_art/relocation_willingness: TEXT[]
-- con CHECK de pertenencia a la lista cerrada, sin tabla de catálogo
-- (son 5 valores fijos, no van a crecer).

ALTER TABLE profiles ADD COLUMN relationship_goals TEXT[] CHECK (relationship_goals <@ ARRAY[
    'casual', 'long_term', 'friendship', 'marriage', 'not_sure'
]::TEXT[]);

-- Migrar el dato existente: quien ya tenía un único valor, pasa a
-- tener un array de un elemento (no se pierde información).
UPDATE profiles SET relationship_goals = ARRAY[relationship_goal] WHERE relationship_goal IS NOT NULL;

ALTER TABLE profiles DROP CONSTRAINT IF EXISTS profiles_relationship_goal_check;
ALTER TABLE profiles DROP COLUMN relationship_goal;
