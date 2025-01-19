-- name: InsertBooks :many
INSERT INTO
    Books (
        ISBN,
        Title,
        Author,
        Publisher,
        Publication_year,
        Volume,
        image_url,
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
        $8
    )
ON CONFLICT (ISBN) DO NOTHING
RETURNING
    ID;

-- name: InsertLibsBooks :many
INSERT INTO
    Libsbooks (
        lib_code,
        isbn,
        class_num,
        scrap
    )
VALUES ($1, $2, $3, $4)
ON CONFLICT (isbn, lib_code) DO NOTHING
RETURNING
    ID;

-- name: InsertLibraries :copyfrom
INSERT INTO
    Libraries (
        lib_code,
        lib_name,
        address,
        Tel,
        Latitude,
        Longitude,
        Homepage,
        Closed,
        operating_time
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
    );

-- name: InsertEmbeddings :exec

INSERT INTO
    Bookembedding (isbn, embedding)
VALUES ($1, $2)
ON CONFLICT (isbn) DO NOTHING;