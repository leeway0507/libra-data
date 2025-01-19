-- +goose Up
-- +goose StatementBegin
ALTER TABLE Books
ADD VectorSearch BOOLEAN DEFAULT False
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE Books
DROP VectorSearch
-- +goose StatementEnd