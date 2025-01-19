-- name: GetBooks :many
SELECT * FROM Books;

-- name: GetBooksFromIsbn :one
SELECT * FROM Books WHERE isbn = $1;

-- name: GetBooksWithoutToc :many
SELECT isbn
FROM Books
WHERE (source IN ('naver', 'daum'))
    AND (
        ISBN LIKE '97911%'
        OR ISBN LIKE '97889%'
        OR ISBN LIKE '97989%'
    )
    AND toc IS NULL
ORDER BY isbn ASC;

-- name: GetBooksWithoutSource :many
SELECT isbn
FROM Books
WHERE
    source is null
    AND (
        ISBN LIKE '97911%'
        OR ISBN LIKE '97889%'
        OR ISBN LIKE '97989%'
    )
ORDER BY isbn ASC
LIMIT 11000;

-- name: ExtractBooksForEmbedding :many
SELECT
    isbn,
    title,
    description,
    toc,
    recommendation
FROM books b
WHERE (
        b.vector_search is false
        and b.source is not null
        AND (
            ISBN LIKE '97911%'
            OR ISBN LIKE '97889%'
            OR ISBN LIKE '97989%'
        )
    )
LIMIT 30000;
-- name: GetLibCodFromLibName :one
SELECT lib_code FROM libraries WHERE lib_name = $1;

-- name: SearchFromBooks :many
SELECT *
FROM books
WHERE
    author LIKE '%$1%'
    OR title LIKE '%$1%'
ORDER BY (
        (
            bigm_similarity (author, $1) + bigm_similarity (title, $1)
        ) * 10
    ) DESC
limit 50;