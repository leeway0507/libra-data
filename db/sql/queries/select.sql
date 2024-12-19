-- name: GetBooks :many
SELECT * FROM Books;

-- name: GetBooksFromIsbn :one
SELECT * FROM Books WHERE isbn = $1;

-- name: ExtractBooksForEmbedding :many
SELECT
    isbn,
    title,
    description,
    toc,
    recommendation
FROM books b
WHERE (
        b.vectorsearch is false
        and b.source is not null
    );
-- name: GetLibCodFromLibName :one
SELECT libcode FROM libraries WHERE libname = $1;