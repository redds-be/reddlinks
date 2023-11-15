-- name: CreateLink :one
INSERT INTO links (id, created_at, expire_at, url)
VALUES ($1, $2, $3, $4)
RETURNING *;
