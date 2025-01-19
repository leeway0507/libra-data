-- +goose Up
-- +goose StatementBegin
ALTER TABLE libraries DROP COLUMN book_count;

ALTER TABLE Libraries RENAME COLUMN lib_address TO address;

ALTER TABLE Libraries RENAME COLUMN longtitude TO longitude;

ALTER TABLE books DROP COLUMN Set_ISBN;

ALTER TABLE libsbooks DROP CONSTRAINT libsbooks_libcode_fkey;

ALTER TABLE libraries
ALTER COLUMN lib_code TYPE varchar(20) USING lib_code::varchar(20);

ALTER TABLE libsbooks
ALTER COLUMN lib_code TYPE varchar(20) USING lib_code::varchar(20);

ALTER TABLE libsbooks
ADD CONSTRAINT libsbooks_libcode_fkey FOREIGN KEY (lib_code) REFERENCES libraries (lib_code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE libraries ADD COLUMN book_count INTEGER;

ALTER TABLE Libraries RENAME COLUMN address TO lib_address;

ALTER TABLE Libraries RENAME COLUMN longitude TO longtitude;

ALTER TABLE books ADD Set_ISBN VARCHAR(255);

ALTER TABLE libsbooks DROP CONSTRAINT libsbooks_libcode_fkey;

ALTER TABLE libraries
ALTER COLUMN lib_code TYPE INTEGER USING lib_code::INTEGER;

ALTER TABLE libsbooks
ALTER COLUMN lib_code TYPE INTEGER USING lib_code::INTEGER;

ALTER TABLE libsbooks
ADD CONSTRAINT libsbooks_libcode_fkey FOREIGN KEY (lib_code) REFERENCES libraries (lib_code);
-- +goose StatementEnd