-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_bigm;
CREATE INDEX title_bigm_idx ON books USING gin (title gin_bigm_ops);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop INDEX title_bigm_idx
-- +goose StatementEnd
