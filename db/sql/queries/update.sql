-- name: UpdateScrapData :exec
UPDATE Books
SET
    Description = $1,
    Recommendation = $2,
    Toc = $3,
    Source = $4,
    Url = $5
WHERE
    isbn = $6;

-- name: UpdateVectorSearchStatus :exec

UPDATE books SET vector_search = $1 WHERE isbn = $2;