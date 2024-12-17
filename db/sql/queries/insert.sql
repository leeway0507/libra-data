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

-- name: InsertEmbeddings :exec

INSERT INTO
    Bookembedding (isbn, embedding)
VALUES ($1, $2)
ON CONFLICT (isbn) DO NOTHING;