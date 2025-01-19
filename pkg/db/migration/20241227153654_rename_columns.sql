-- +goose Up
-- +goose StatementBegin
ALTER TABLE Books
RENAME COLUMN PublicationYear TO publication_year;
ALTER TABLE Books
RENAME COLUMN ImageURL TO image_url;
ALTER TABLE Books
RENAME COLUMN SetISBN TO set_isbn;
ALTER TABLE Books
RENAME COLUMN VectorSearch TO vector_search;

ALTER TABLE Libraries
RENAME COLUMN LibCode TO lib_code;
ALTER TABLE Libraries
RENAME COLUMN LibName TO lib_name;
ALTER TABLE Libraries
RENAME COLUMN LibAddress TO lib_address;
ALTER TABLE Libraries
RENAME COLUMN OperatingTime TO operating_time;
ALTER TABLE Libraries
RENAME COLUMN BookCount TO book_count;

ALTER TABLE LibsBooks
RENAME COLUMN LibCode TO lib_code;
ALTER TABLE LibsBooks
RENAME COLUMN ClassNum TO class_num;
ALTER TABLE LibsBooks
RENAME COLUMN BookCode TO book_code;
ALTER TABLE LibsBooks
RENAME COLUMN ShelfCode TO shelf_code;
ALTER TABLE LibsBooks
RENAME COLUMN ShelfName TO shelf_name;



-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE books
RENAME COLUMN publication_year TO PublicationYear;
ALTER TABLE books
RENAME COLUMN image_url TO ImageURL;
ALTER TABLE books
RENAME COLUMN set_isbn TO SetISBN;
ALTER TABLE books
RENAME COLUMN vector_search TO VectorSearch;

ALTER TABLE Libraries
RENAME COLUMN lib_code TO LibCode;
ALTER TABLE Libraries
RENAME COLUMN lib_name TO LibName;
ALTER TABLE Libraries
RENAME COLUMN lib_address TO LibAddress;
ALTER TABLE Libraries
RENAME COLUMN operating_time TO OperatingTime;
ALTER TABLE Libraries
RENAME COLUMN book_count TO BookCount;

ALTER TABLE LibsBooks
RENAME COLUMN lib_code TO LibCode;
ALTER TABLE LibsBooks
RENAME COLUMN class_num TO ClassNum;
ALTER TABLE LibsBooks
RENAME COLUMN book_code TO BookCode;
ALTER TABLE LibsBooks
RENAME COLUMN shelf_code TO ShelfCode;
ALTER TABLE LibsBooks
RENAME COLUMN shelf_name TO ShelfName;
-- +goose StatementEnd
