-- +goose Up
-- +goose StatementBegin
ALTER TABLE Books
ADD Toc TEXT,
ADD Recommendation TEXT,
ADD ScrapSource VARCHAR(50),
ADD ScrapUrl VARCHAR(512);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE Books
DROP Toc,
DROP ScrapSource,
DROP ScrapUrl,
DROP Recommendation;
-- +goose StatementEnd