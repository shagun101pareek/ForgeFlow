ALTER TABLE generations
    ADD COLUMN status TEXT NOT NULL DEFAULT 'completed',
    ADD COLUMN error_message TEXT NOT NULL DEFAULT '';

ALTER TABLE generations
    ALTER COLUMN status SET DEFAULT 'queued';
