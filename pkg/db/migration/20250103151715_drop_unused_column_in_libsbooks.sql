-- +goose Up
-- +goose StatementBegin
ALTER TABLE libsbooks DROP COLUMN book_code;

ALTER TABLE libsbooks DROP COLUMN shelf_code;

ALTER TABLE libsbooks DROP COLUMN shelf_name;

ALTER TABLE libsbooks ADD COLUMN scrap BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE libsbooks ADD COLUMN book_code VARCHAR(100);

ALTER TABLE libsbooks ADD COLUMN shelf_code VARCHAR(100);

ALTER TABLE libsbooks ADD COLUMN shelf_name VARCHAR(100);

ALTER TABLE libsbooks DROP COLUMN scrap;
-- +goose StatementEnd