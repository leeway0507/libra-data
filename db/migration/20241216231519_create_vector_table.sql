-- +goose Up
-- +goose StatementBegin

-- CREATE EXTENSION vector; <= need to exec at cli(psql library_search and then exec this query)
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE BookEmbedding (
    ID SERIAL PRIMARY KEY,
    ISBN VARCHAR(15),
    embedding vector (1536),
    FOREIGN KEY (ISBN) REFERENCES Books (ISBN)
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
CREATE EXTENSION vector;

DROP TABLE IF EXISTS BookEmbedding;
-- +goose StatementEnd