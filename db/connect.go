package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func connectPGPool(url string, ctx *context.Context) *pgxpool.Pool {

	fmt.Println("trying to connect to db : ", url)
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		panic(err)
	}
	pool, err := pgxpool.NewWithConfig(*ctx, config)
	if err != nil {
		panic(err)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return pool

}
func connectPG(url string, ctx context.Context) *pgx.Conn {

	fmt.Println("trying to connect to db : ", url)
	conn, err := pgx.Connect(ctx, url)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return conn

}
