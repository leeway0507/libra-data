-- name: GetBooks :many
SELECT * FROM Books;

-- name: InsertBooks :copyfrom
INSERT INTO
    Books (
        ISBN,
        Title,
        Author,
        Publisher,
        PublicationYear,
        SetISBN,
        AdditionalCode,
        Volume,
        SubjectCode,
        BookCount,
        LoanCount,
        RegistrationDate
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
        $10,
        $11,
        $12
    );

-- name: InsertLibraries :copyfrom
INSERT INTO
    libraries (
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