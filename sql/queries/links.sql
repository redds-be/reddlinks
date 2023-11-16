-- name: CreateLink :one
INSERT INTO links (id, created_at, expire_at, url, short)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetLinkByShort :one
SELECT * FROM links WHERE short = $1;

-- name: GetLinks :many
SELECT * FROM links;

-- name: RemoveLink :exec
DELETE FROM links WHERE short = $1;