CREATE TABLE generations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    specification JSONB NOT NULL,
    files JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX generations_project_id_created_at_idx ON generations (project_id, created_at DESC);
