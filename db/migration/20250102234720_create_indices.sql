-- +goose Up
-- +goose StatementBegin
SET LOCAL maintenance_work_mem = '256MB';

CREATE INDEX book_embedding_ivfflat ON bookembedding USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 600);

SET ivfflat.probes = 10;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SET LOCAL maintenance_work_mem = '64MB';

DROP INDEX book_embedding_ivfflat;

SET ivfflat.probes = 1;
-- +goose StatementEnd