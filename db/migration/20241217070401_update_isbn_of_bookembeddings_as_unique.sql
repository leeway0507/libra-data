-- +goose Up
-- +goose StatementBegin

ALTER TABLE bookembedding
DROP COLUMN isbn,
ADD COLUMN isbn VARCHAR(15) NOT NULL UNIQUE;
-- ADD CONSTRAINT bookembedding_isbn UNIQUE (isbn);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE bookembedding
DROP COLUMN isbn,
ADD COLUMN isbn VARCHAR(15)
-- +goose StatementEnd