package db

import (
	"context"
	"libraData/config"
	sqlc "libraData/db/sqlc"
	"testing"
)

func Test_Insert(t *testing.T) {
	config.SetTestEnvConfig(cfg)

	ctx := context.Background()
	conn := ConnectPG(cfg.DATABASE_TEST_URL, ctx)
	defer conn.Close(ctx)

	testQuery := sqlc.New(conn)

	t.Run("init test db", func(t *testing.T) {
		defer func() {
			if _, err := testQuery.DeleteAllBooksRowForTest(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := testQuery.DeleteAllLibariesForTest(ctx); err != nil {
				t.Fatal(err)
			}
		}()
	})

}
