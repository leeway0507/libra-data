-- name: SearchQuery :many
SELECT b.title,b.author, embedding <=> $1 as sim
FROM BookEmbedding e
JOIN books b
ON b.isbn = e.isbn
ORDER BY embedding <=> $1 ASC
LIMIT 50;
