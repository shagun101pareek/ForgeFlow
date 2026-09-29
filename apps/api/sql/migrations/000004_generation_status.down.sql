ALTER TABLE generations
    DROP COLUMN IF EXISTS error_message,
    DROP COLUMN IF EXISTS status;
