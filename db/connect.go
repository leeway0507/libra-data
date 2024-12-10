package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func connectPG(url string, ctx *context.Context) *pgx.Conn {

	fmt.Println("trying to connect to db : ", url)
	conn, err := pgx.Connect(*ctx, url)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return conn

}
