-- name: GetBooks :many
SELECT * FROM Books;

-- name: InsertBooks :many
INSERT INTO
    Books (
        ISBN,
        Title,
        Author,
        Publisher,
        PublicationYear,
        SetISBN,
        Volume,
        ImageURL,
        Description
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9
    )
ON CONFLICT (ISBN) DO NOTHING
RETURNING
    ID;
-- name: UpdateScrapResult :exec
UPDATE Books
SET
    Description = $1,
    Recommendation = $2,
    Toc = $3,
    Source = $4,
    Url = $5
WHERE
    isbn = $6;
-- name: InsertLibsBooks :many
INSERT INTO
    Libsbooks (
        libcode,
        isbn,
        classnum,
        bookcode,
        shelfcode,
        shelfname
    )
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (isbn, libcode) DO NOTHING
RETURNING
    ID;

-- name: DeleteAllBooksRowForTest :one
DELETE FROM Books RETURNING ID;

-- name: DeleteAllLibariesForTest :one
DELETE FROM Libraries RETURNING ID;

-- name: DeleteAllLibsBooksForTest :one
DELETE FROM libsbooks RETURNING ID;

-- name: InsertLibraries :copyfrom
INSERT INTO
    Libraries (
        LibCode,
        LibName,
        LibAddress,
        Tel,
        Latitude,
        Longtitude,
        Homepage,
        Closed,
        OperatingTime,
        BookCount
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10
    );