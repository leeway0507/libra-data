-- name: InsertBooks :many
INSERT INTO
    Books (
        ISBN,
        Title,
        Author,
        Publisher,
        Publication_year,
        set_isbn,
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
        $8,
        $9
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
        book_code,
        shelf_code,
        shelf_name
    )
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (isbn, lib_code) DO NOTHING
RETURNING
    ID;

-- name: InsertLibraries :copyfrom
INSERT INTO
    Libraries (
        lib_code,
        lib_name,
        lib_address,
        Tel,
        Latitude,
        Longtitude,
        Homepage,
        Closed,
        operating_time,
        book_count
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