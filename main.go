package libraData

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func libraData() {
	cfg := GetEnvConfig()

	fmt.Println("trying to connect to db : ", cfg)
	conn, err := pgx.Connect(context.Background(), cfg.DATABASE_URL)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// fmt.Println(name, weight)
}
