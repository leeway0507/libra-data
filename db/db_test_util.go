package db

import (
	"context"
	"io"
	sqlc "libraData/db/sqlc"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
)

var sqlPath string = "/Users/yangwoolee/repo/libra-data/db/sql/schema.sql"
var testPath string = "/Users/yangwoolee/repo/libra-data/data/test"

func InitTestTable(conn *pgx.Conn, ctx context.Context) error {
	testQuery := sqlc.New(conn)
	err := DropTestTable(conn, ctx)
	if err != nil {
		return err
	}
	err = CreateTestTable(conn, ctx)
	if err != nil {
		return err
	}
	err = InsertLibBookBulkFromJSON(testQuery, ctx, filepath.Join(testPath, "insert-books.json"))
	if err != nil {
		return err
	}
	err = InsertLibInfoBulkFromJSON(testQuery, ctx, filepath.Join(testPath, "libinfo-test.json"))
	if err != nil {
		return err
	}
	err = InsertLibsBooksRelationBulkFromJSON(testQuery, ctx, filepath.Join(testPath, "insert-books.json"), 127058)
	if err != nil {
		return err
	}
	return nil
}

func CreateTestTable(conn *pgx.Conn, ctx context.Context) error {
	file, err := os.Open(sqlPath)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	query := string(b)

	parts := strings.Split(query, "CREATE TABLE")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			createQuery := "CREATE TABLE " + trimmed
			_, err = conn.Exec(ctx, createQuery)
			if err != nil {
				return err
			}
			// fmt.Println("create query", createQuery)
		}
	}
	return nil
}

func DropTestTable(conn *pgx.Conn, ctx context.Context) error {
	file, err := os.Open(sqlPath)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	query := string(b)

	parts := strings.Split(query, "CREATE TABLE")

	for idx := range parts {
		trimmed := strings.Split(parts[len(parts)-(idx+1)], "(")[0]
		if trimmed != "" {
			dropQuery := "DROP TABLE IF EXISTS" + trimmed + ";"
			_, err = conn.Exec(ctx, dropQuery)
			if err != nil {
				return err
			}
			// fmt.Println("drop query : ", dropQuery)
		}
	}
	return nil
}
