-- name: ListProjectsByUser :many
SELECT
    p.id,
    p.name,
    p.description,
    p.created_at,
    p.updated_at,
    (p.image IS NOT NULL)::boolean AS has_image,
    COALESCE(latest.prompt, '') AS latest_prompt,
    COALESCE(latest.status, '') AS latest_status
FROM projects p
LEFT JOIN LATERAL (
    SELECT g.prompt, g.status
    FROM generations g
    WHERE g.project_id = p.id
    ORDER BY g.created_at DESC
    LIMIT 1
) latest ON true
WHERE p.user_id = $1
ORDER BY p.updated_at DESC;

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
