-- +goose Up
-- +goose StatementBegin
CREATE TABLE LibsBooks (
    ID SERIAL PRIMARY KEY,
    LibCode INTEGER,
    ISBN VARCHAR(15),
    FOREIGN KEY (LibCode) REFERENCES Libraries (LibCode) ON DELETE CASCADE,
    FOREIGN KEY (ISBN) REFERENCES Books (ISBN) ON DELETE CASCADE,
    UNIQUE (LibCode, ISBN)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE LibsBooks;
-- +goose StatementEnd