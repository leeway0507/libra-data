-- +goose Up
-- +goose StatementBegin
CREATE TABLE Books (
    ID SERIAL PRIMARY KEY,
    ISBN VARCHAR(15) UNIQUE,
    Title VARCHAR(1024),
    Author VARCHAR(255),
    Publisher VARCHAR(255),
    PublicationYear VARCHAR(50),
    SetISBN VARCHAR(255),
    AdditionalCode VARCHAR(255),
    Volume VARCHAR(50)
);

CREATE TABLE Libraries (
    ID SERIAL PRIMARY KEY,
    LibCode INTEGER UNIQUE,
    LibName VARCHAR(100),
    LibAddress VARCHAR(255),
    Tel VARCHAR(100),
    Latitude FLOAT,
    Longtitude FLOAT,
    Homepage VARCHAR(100),
    Closed VARCHAR(512),
    OperatingTime VARCHAR(512),
    BookCount INTEGER
);

CREATE TABLE LibsBooks (
    ID SERIAL PRIMARY KEY,
    LibCode INTEGER,
    ISBN VARCHAR(15),
    CLASSNUM VARCHAR(255),
    BOOKCODE VARCHAR(100),
    SHELFCODE VARCHAR(100),
    SHELFNAME VARCHAR(100),
    FOREIGN KEY (LibCode) REFERENCES Libraries (LibCode) ON DELETE CASCADE,
    FOREIGN KEY (ISBN) REFERENCES Books (ISBN) ON DELETE CASCADE,
    UNIQUE (LibCode, ISBN)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Books;

DROP TABLE Libraries;

DROP TABLE LibsBooks;
-- +goose StatementEnd