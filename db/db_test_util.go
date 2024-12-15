package db

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
)

var sqlPath string = "/Users/yangwoolee/repo/libra-data/db/sql/schema.sql"

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
			fmt.Println("create query", createQuery)
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
			fmt.Println("drop query : ", dropQuery)
		}
	}
	return nil
}
