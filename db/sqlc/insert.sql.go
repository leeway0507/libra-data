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
        PublicationYear,
        SetISBN,
        Volume,
        ImageURL,
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
        $8,
        $9
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
	Publicationyear pgtype.Text
	Setisbn         pgtype.Text
	Volume          pgtype.Text
	Imageurl        pgtype.Text
	Description     pgtype.Text
}

func (q *Queries) InsertBooks(ctx context.Context, arg InsertBooksParams) ([]int32, error) {
	rows, err := q.db.Query(ctx, insertBooks,
		arg.Isbn,
		arg.Title,
		arg.Author,
		arg.Publisher,
		arg.Publicationyear,
		arg.Setisbn,
		arg.Volume,
		arg.Imageurl,
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
	Libcode       pgtype.Int4
	Libname       pgtype.Text
	Libaddress    pgtype.Text
	Tel           pgtype.Text
	Latitude      pgtype.Float8
	Longtitude    pgtype.Float8
	Homepage      pgtype.Text
	Closed        pgtype.Text
	Operatingtime pgtype.Text
	Bookcount     pgtype.Int4
}

const insertLibsBooks = `-- name: InsertLibsBooks :many
INSERT INTO
    Libsbooks (
        libcode,
        isbn,
        classnum,
        bookcode,
        shelfcode,
        shelfname
    )
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (isbn, libcode) DO NOTHING
RETURNING
    ID
`

type InsertLibsBooksParams struct {
	Libcode   pgtype.Int4
	Isbn      pgtype.Text
	Classnum  pgtype.Text
	Bookcode  pgtype.Text
	Shelfcode pgtype.Text
	Shelfname pgtype.Text
}

func (q *Queries) InsertLibsBooks(ctx context.Context, arg InsertLibsBooksParams) ([]int32, error) {
	rows, err := q.db.Query(ctx, insertLibsBooks,
		arg.Libcode,
		arg.Isbn,
		arg.Classnum,
		arg.Bookcode,
		arg.Shelfcode,
		arg.Shelfname,
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
