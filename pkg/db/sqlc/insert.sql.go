// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: insert.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
)

const insertBooks = `-- name: InsertBooks :many
INSERT INTO
    Books (
        ISBN,
        Title,
        Author,
        Publisher,
        Publication_year,
        Volume,
        image_url,
        Description
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
    )
ON CONFLICT (ISBN) DO NOTHING
RETURNING
    ID
`

type InsertBooksParams struct {
	Isbn            pgtype.Text
	Title           pgtype.Text
	Author          pgtype.Text
	Publisher       pgtype.Text
	PublicationYear pgtype.Text
	Volume          pgtype.Text
	ImageUrl        pgtype.Text
	Description     pgtype.Text
}

func (q *Queries) InsertBooks(ctx context.Context, arg InsertBooksParams) ([]int32, error) {
	rows, err := q.db.Query(ctx, insertBooks,
		arg.Isbn,
		arg.Title,
		arg.Author,
		arg.Publisher,
		arg.PublicationYear,
		arg.Volume,
		arg.ImageUrl,
		arg.Description,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertEmbeddings = `-- name: InsertEmbeddings :exec

INSERT INTO
    Bookembedding (isbn, embedding)
VALUES ($1, $2)
ON CONFLICT (isbn) DO NOTHING
`

type InsertEmbeddingsParams struct {
	Isbn      string
	Embedding pgvector.Vector
}

func (q *Queries) InsertEmbeddings(ctx context.Context, arg InsertEmbeddingsParams) error {
	_, err := q.db.Exec(ctx, insertEmbeddings, arg.Isbn, arg.Embedding)
	return err
}

type InsertLibrariesParams struct {
	LibCode       pgtype.Text
	LibName       pgtype.Text
	Address       pgtype.Text
	Tel           pgtype.Text
	Latitude      pgtype.Float8
	Longitude     pgtype.Float8
	Homepage      pgtype.Text
	Closed        pgtype.Text
	OperatingTime pgtype.Text
}

const insertLibsBooks = `-- name: InsertLibsBooks :many
INSERT INTO
    Libsbooks (
        lib_code,
        isbn,
        class_num,
        scrap
    )
VALUES ($1, $2, $3, $4)
ON CONFLICT (isbn, lib_code) DO NOTHING
RETURNING
    ID
`

type InsertLibsBooksParams struct {
	LibCode  pgtype.Text
	Isbn     pgtype.Text
	ClassNum pgtype.Text
	Scrap    pgtype.Bool
}

func (q *Queries) InsertLibsBooks(ctx context.Context, arg InsertLibsBooksParams) ([]int32, error) {
	rows, err := q.db.Query(ctx, insertLibsBooks,
		arg.LibCode,
		arg.Isbn,
		arg.ClassNum,
		arg.Scrap,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
