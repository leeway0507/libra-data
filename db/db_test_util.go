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
			fmt.Println("create query", createQuery)
			if err != nil {
				fmt.Println(err.Error())
				return err
			}
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
			dropQuery := "DROP TABLE " + trimmed + ";"
			_, err = conn.Exec(ctx, dropQuery)
			fmt.Println("drop query : ", dropQuery)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
