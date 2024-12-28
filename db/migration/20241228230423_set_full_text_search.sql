-- +goose Up
-- +goose StatementBegin
ALTER TABLE books
 ADD COLUMN document tsvector;

UPDATE books
SET document = to_tsvector(title || ' ' || author || ' ' ||toc);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE books
DROP COLUMN document;
-- +goose StatementEnd
