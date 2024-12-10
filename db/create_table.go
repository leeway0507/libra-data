package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func Create_Book_Table(conn *pgx.Conn, ctx *context.Context) error {
	query := `CREATE TABLE Book (
				ID SERIAL PRIMARY KEY,
				ISBN VARCHAR(15) UNIQUE,
				Title VARCHAR(1024),
				Author VARCHAR(255),
				Publisher VARCHAR(255),
				PublicationYear VARCHAR(50),
				SetISBN VARCHAR(255),
				AdditionalCode VARCHAR(255),
				Volume VARCHAR(50),
				SubjectCode VARCHAR(50),
				BookCount INTEGER DEFAULT 0,
				LoanCount INTEGER DEFAULT 0,
				RegistrationDate DATE
			)
	`
	_, err := conn.Query(*ctx, query)
	if err != nil {
		return err
	}
	return nil
}
