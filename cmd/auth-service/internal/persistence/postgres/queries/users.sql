-- name: CreateUser :exec
INSERT INTO users (id, email, phone, role, google_id)
VALUES ($1, $2, $3, $4, $5);

-- name: FindByEmail :one
SELECT id, email, phone, role, google_id
FROM users
WHERE email = $1;

-- name: FindByPhone :one
SELECT id, email, phone, role, google_id
FROM users
WHERE phone = $1;
