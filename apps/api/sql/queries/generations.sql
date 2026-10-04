-- name: SaveGeneration :one
INSERT INTO generations (project_id, prompt, specification, files, status)
VALUES ($1, $2, $3, $4, 'completed')
RETURNING id, project_id, prompt, specification, files, created_at, status, error_message, github_url;

-- name: CreateGeneration :one
INSERT INTO generations (project_id, prompt, specification, files, status)
VALUES ($1, $2, '{}'::jsonb, '[]'::jsonb, 'queued')
RETURNING id, project_id, prompt, specification, files, created_at, status, error_message, github_url;

-- name: SetGenerationStatus :exec
UPDATE generations
SET status = $2
WHERE id = $1 AND status = 'queued';

-- name: CompleteGeneration :exec
UPDATE generations
SET status = 'completed',
    specification = $2,
    files = $3,
    error_message = ''
WHERE id = $1 AND status = 'running';

-- name: FailGeneration :exec
UPDATE generations
SET status = 'failed',
    error_message = $2
WHERE id = $1 AND status IN ('queued', 'running');

-- name: FailStaleGenerations :exec
UPDATE generations
SET status = 'failed',
    error_message = 'Generation timed out'
WHERE project_id = $1
  AND status IN ('queued', 'running')
  AND created_at < now() - interval '2 minutes';

-- name: SetGenerationGitHubURL :exec
UPDATE generations
SET github_url = $2
WHERE id = $1;

-- name: GetLatestGenerationForUserProject :one
SELECT g.id, g.project_id, g.prompt, g.specification, g.files, g.created_at, g.status, g.error_message, g.github_url
FROM generations g
JOIN projects p ON p.id = g.project_id
WHERE g.project_id = $1 AND p.user_id = $2
ORDER BY g.created_at DESC
LIMIT 1;

-- name: ListGenerationsForUserProject :many
SELECT g.id, g.prompt, g.status, g.created_at
FROM generations g
JOIN projects p ON p.id = g.project_id
WHERE g.project_id = $1 AND p.user_id = $2
ORDER BY g.created_at DESC;

-- name: DeleteGenerationForUser :execrows
DELETE FROM generations
WHERE generations.id = $1
  AND generations.project_id = $2
  AND EXISTS (
    SELECT 1 FROM projects
    WHERE projects.id = generations.project_id AND projects.user_id = $3
  );

-- name: GetGenerationForUser :one
SELECT g.id, g.project_id, g.prompt, g.specification, g.files, g.created_at, g.status, g.error_message, g.github_url
FROM generations g
JOIN projects p ON p.id = g.project_id
WHERE g.id = $1 AND g.project_id = $2 AND p.user_id = $3;
