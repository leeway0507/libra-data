CREATE TABLE Books (
    ID SERIAL PRIMARY KEY,
    ISBN VARCHAR(15) UNIQUE,
    Title VARCHAR(1024),
    Author VARCHAR(512),
    Publisher VARCHAR(255),
    PublicationYear VARCHAR(50),
    SetISBN VARCHAR(255),
    Volume VARCHAR(50),
    ImageURL VARCHAR(512),
    BookDescription TEXT,
    Recommendation TEXT,
    Toc TEXT,
    ScrapSource VARCHAR(50),
    ScrapUrl VARCHAR(512)
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
    ClassNum VARCHAR(255),
    BookCode VARCHAR(100),
    ShelfCode VARCHAR(100),
    ShelfName VARCHAR(100),
    FOREIGN KEY (LibCode) REFERENCES Libraries (LibCode) ON DELETE CASCADE,
    FOREIGN KEY (ISBN) REFERENCES Books (ISBN) ON DELETE CASCADE,
    UNIQUE (LibCode, ISBN)
);