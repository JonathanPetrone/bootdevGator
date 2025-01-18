-- name: GetUser :one
SELECT * 
FROM users
WHERE users.name = $1;