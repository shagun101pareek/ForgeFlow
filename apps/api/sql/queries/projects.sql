-- name: ListProjectsByUser :many
SELECT id, name, description, created_at, updated_at
FROM projects
WHERE user_id = $1
ORDER BY updated_at DESC;

-- name: CreateProject :one
INSERT INTO projects (user_id, name, description)
VALUES ($1, $2, $3)
RETURNING id, name, description, created_at, updated_at;

-- name: GetProjectByIDForUser :one
SELECT id, name, description, created_at, updated_at
FROM projects
WHERE id = $1 AND user_id = $2;

-- name: UpdateProject :one
UPDATE projects
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    updated_at = now()
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING id, name, description, created_at, updated_at;

-- name: DeleteProject :execrows
DELETE FROM projects
WHERE id = $1 AND user_id = $2;
