-- 000011_add_likes_and_matches.up.sql

CREATE TABLE likes (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    to_profile_id   UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT likes_distinct_profiles CHECK (from_profile_id <> to_profile_id),
    UNIQUE (from_profile_id, to_profile_id)
);

CREATE INDEX likes_from_profile_id_idx ON likes (from_profile_id);
CREATE INDEX likes_to_profile_id_idx ON likes (to_profile_id);

CREATE TABLE matches (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_one_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    profile_two_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT matches_distinct_profiles CHECK (profile_one_id <> profile_two_id),
    CONSTRAINT matches_ordered_pair CHECK (profile_one_id < profile_two_id),
    UNIQUE (profile_one_id, profile_two_id)
);

CREATE INDEX matches_profile_one_id_idx ON matches (profile_one_id);
CREATE INDEX matches_profile_two_id_idx ON matches (profile_two_id);
