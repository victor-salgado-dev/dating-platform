-- Índices para resolver la consulta de favoritos mutuos con Index Scan.
--
-- favorites.user_id            -> users.id
-- favorites.favorite_profile_id -> profiles.id
--
-- idx_favorites_user_target: cubre "a quién marqué como favorito"
--   (sent) y el lado directo del JOIN de mutualidad.
-- idx_favorites_target: cubre el JOIN inverso "quién me marcó a mí"
--   (received / lado back del mutual).

CREATE INDEX IF NOT EXISTS idx_favorites_user_target
  ON favorites(user_id, favorite_profile_id);

CREATE INDEX IF NOT EXISTS idx_favorites_target
  ON favorites(favorite_profile_id);
