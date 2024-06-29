-- name: CreateUser :exec
INSERT INTO users (id, email, phone, role, password)
VALUES ($1, $2, $3, $4, $5);

-- name: FindUserByEmail :one
SELECT id, email, phone, role, password
FROM users
WHERE email = $1;

-- name: FindUserByPhone :one
SELECT id, email, phone, role, password
FROM users
WHERE phone = $1;
