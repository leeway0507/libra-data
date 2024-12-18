// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: select.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const extractBooksForEmbedding = `-- name: ExtractBooksForEmbedding :many
SELECT
    isbn,
    title,
    description,
    toc,
    recommendation
FROM books b
WHERE (
        b.vectorsearch is false
        and b.source is not null
    )
`

type ExtractBooksForEmbeddingRow struct {
	Isbn           pgtype.Text
	Title          pgtype.Text
	Description    pgtype.Text
	Toc            pgtype.Text
	Recommendation pgtype.Text
}

func (q *Queries) ExtractBooksForEmbedding(ctx context.Context) ([]ExtractBooksForEmbeddingRow, error) {
	rows, err := q.db.Query(ctx, extractBooksForEmbedding)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ExtractBooksForEmbeddingRow
	for rows.Next() {
		var i ExtractBooksForEmbeddingRow
		if err := rows.Scan(
			&i.Isbn,
			&i.Title,
			&i.Description,
			&i.Toc,
			&i.Recommendation,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getBooks = `-- name: GetBooks :many
SELECT id, isbn, title, author, publisher, publicationyear, setisbn, volume, imageurl, description, recommendation, toc, source, url, vectorsearch FROM Books
`

func (q *Queries) GetBooks(ctx context.Context) ([]Book, error) {
	rows, err := q.db.Query(ctx, getBooks)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Book
	for rows.Next() {
		var i Book
		if err := rows.Scan(
			&i.ID,
			&i.Isbn,
			&i.Title,
			&i.Author,
			&i.Publisher,
			&i.Publicationyear,
			&i.Setisbn,
			&i.Volume,
			&i.Imageurl,
			&i.Description,
			&i.Recommendation,
			&i.Toc,
			&i.Source,
			&i.Url,
			&i.Vectorsearch,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getBooksFromIsbn = `-- name: GetBooksFromIsbn :one
SELECT id, isbn, title, author, publisher, publicationyear, setisbn, volume, imageurl, description, recommendation, toc, source, url, vectorsearch FROM Books WHERE isbn = $1
`

func (q *Queries) GetBooksFromIsbn(ctx context.Context, isbn pgtype.Text) (Book, error) {
	row := q.db.QueryRow(ctx, getBooksFromIsbn, isbn)
	var i Book
	err := row.Scan(
		&i.ID,
		&i.Isbn,
		&i.Title,
		&i.Author,
		&i.Publisher,
		&i.Publicationyear,
		&i.Setisbn,
		&i.Volume,
		&i.Imageurl,
		&i.Description,
		&i.Recommendation,
		&i.Toc,
		&i.Source,
		&i.Url,
		&i.Vectorsearch,
	)
	return i, err
}
