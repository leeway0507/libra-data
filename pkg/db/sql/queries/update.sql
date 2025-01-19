-- name: UpdateScrapData :exec
UPDATE Books
SET
    Description = $1,
    Recommendation = $2,
    Toc = $3,
    Source = $4,
    Url = $5,
    image_url = $6
WHERE
    isbn = $7;

-- name: UpdateVectorSearchStatus :exec
UPDATE books SET vector_search = $1 WHERE isbn = $2;

-- name: UpdateToc :exec
UPDATE books SET toc = $1 WHERE isbn = $2;