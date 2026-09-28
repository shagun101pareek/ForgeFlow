-- name: CreateGeneration :one
INSERT INTO generations (project_id, prompt, specification, files)
VALUES ($1, $2, $3, $4)
RETURNING id, project_id, prompt, specification, files, created_at;

-- name: GetLatestGenerationForUserProject :one
SELECT g.id, g.project_id, g.prompt, g.specification, g.files, g.created_at
FROM generations g
JOIN projects p ON p.id = g.project_id
WHERE g.project_id = $1 AND p.user_id = $2
ORDER BY g.created_at DESC
LIMIT 1;

-- name: ListGenerationsForUserProject :many
SELECT g.id, g.prompt, g.created_at
FROM generations g
JOIN projects p ON p.id = g.project_id
WHERE g.project_id = $1 AND p.user_id = $2
ORDER BY g.created_at DESC;

-- name: GetGenerationForUser :one
SELECT g.id, g.project_id, g.prompt, g.specification, g.files, g.created_at
FROM generations g
JOIN projects p ON p.id = g.project_id
WHERE g.id = $1 AND g.project_id = $2 AND p.user_id = $3;
