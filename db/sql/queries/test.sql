-- name: DeleteAllBooksRowForTest :one
DELETE FROM Books RETURNING ID;

-- name: DeleteAllLibariesForTest :one
DELETE FROM Libraries RETURNING ID;

-- name: DeleteAllLibsBooksForTest :one
DELETE FROM libsbooks RETURNING ID;