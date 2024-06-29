-- name: CreateToken :exec
INSERT INTO oauth2_tokens (access_token, refresh_token, expires_at, client_id, user_id)
VALUES ($1, $2, $3, $4, $5);

-- name: GetTokenByAccess :one
SELECT id, access_token, refresh_token, expires_at, client_id, user_id
FROM oauth2_tokens
WHERE access_token = $1;

-- name: GetTokenByRefresh :one
SELECT id, access_token, refresh_token, expires_at, client_id, user_id
FROM oauth2_tokens
WHERE refresh_token = $1;

-- name: UpdateTokenByUserID :exec
UPDATE oauth2_tokens
SET access_token = $1, expires_at = $2
WHERE user_id = $3;

-- name: CreateClient :exec
INSERT INTO oauth2_clients (client_id, client_secret, domain, scope, redirect_url)
VALUES ($1, $2, $3, $4, $5);

-- name: GetClientByID :one
SELECT id, client_id, client_secret, domain, scope, redirect_url
FROM oauth2_clients
WHERE client_id = $1;

-- name: DeleteClientByID :exec
DELETE FROM oauth2_clients
WHERE client_id = $1;

-- name: GetScopeByClientID :one
SELECT scope
FROM oauth2_clients
WHERE client_id = $1;

-- name: GetClientIdByUserId :one
SELECT a.client_id
FROM oauth2_clients a
JOIN oauth2_tokens b ON a.client_id = b.client_id
WHERE b.user_id = $1;