CREATE TABLE profile_visits (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    visitor_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    visited_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    visited_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT profile_visits_distinct_profiles CHECK (visitor_profile_id <> visited_profile_id),
    UNIQUE (visitor_profile_id, visited_profile_id)
);

CREATE INDEX profile_visits_visitor_idx ON profile_visits (visitor_profile_id, visited_at DESC);
CREATE INDEX profile_visits_visited_idx ON profile_visits (visited_profile_id, visited_at DESC);