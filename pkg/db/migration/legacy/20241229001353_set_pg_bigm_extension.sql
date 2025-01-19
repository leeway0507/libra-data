-- https://new-pow.tistory.com/77
-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_bigm ;
CREATE INDEX title_idx ON books USING gin (title gin_bigm_ops);
CREATE INDEX author_idx ON books USING gin (author gin_bigm_ops);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX title_idx;
DROP INDEX author_idx;
-- +goose StatementEnd

