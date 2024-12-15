-- +goose Up
-- +goose StatementBegin
ALTER TABLE Books
DROP BookDescription,
ADD Toc TEXT,
ADD Recommendation TEXT,
ADD Source VARCHAR(50),
ADD Url VARCHAR(512),
ADD Description TEXT;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE Books
DROP Toc,
DROP ScrapSource,
DROP ScrapUrl,
DROP Recommendation;
-- +goose StatementEnd