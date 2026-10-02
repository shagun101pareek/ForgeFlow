-- name: ListProjectsByUser :many
SELECT id, name, description, created_at, updated_at, (image IS NOT NULL)::boolean AS has_image
FROM projects
WHERE user_id = $1
ORDER BY updated_at DESC;

-- name: CreateProject :one
INSERT INTO projects (user_id, name, description)
VALUES ($1, $2, $3)
RETURNING id, name, description, created_at, updated_at, (image IS NOT NULL)::boolean AS has_image;

-- name: GetProjectByIDForUser :one
SELECT id, name, description, created_at, updated_at, (image IS NOT NULL)::boolean AS has_image
FROM projects
WHERE id = $1 AND user_id = $2;

-- name: UpdateProject :one
UPDATE projects
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    updated_at = now()
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id')
RETURNING id, name, description, created_at, updated_at, (image IS NOT NULL)::boolean AS has_image;

-- name: GetProjectImageForUser :one
SELECT image, image_type
FROM projects
WHERE id = $1 AND user_id = $2;

-- name: SetProjectImage :exec
UPDATE projects
SET image = $3, image_type = $4, updated_at = now()
WHERE id = $1 AND user_id = $2;

-- name: ClearProjectImage :execrows
UPDATE projects
SET image = NULL, image_type = '', updated_at = now()
WHERE id = $1 AND user_id = $2;

-- name: DeleteProject :execrows
DELETE FROM projects
WHERE id = $1 AND user_id = $2;
