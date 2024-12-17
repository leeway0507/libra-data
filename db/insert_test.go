package db

import (
	"context"
	"fmt"
	"libraData/config"
	sqlc "libraData/db/sqlc"
	"path/filepath"
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

	t.Run("Insert books into db", func(t *testing.T) {
		err := InsertLibBookBulkFromJSON(testQuery, ctx, filepath.Join(cfg.DATA_PATH, "insert-books.json"))
		if err != nil {
			t.Fatal(err)
		}

	})
	t.Run("Insert libinfos into db", func(t *testing.T) {

		err := InsertLibInfoBulkFromJSON(testQuery, ctx, filepath.Join(cfg.DATA_PATH, "libinfo-test.json"))
		if err != nil {
			if _, err := testQuery.DeleteAllLibariesForTest(ctx); err != nil {
				fmt.Println(err)
			}
			err := InsertLibInfoBulkFromJSON(testQuery, ctx, filepath.Join(cfg.DATA_PATH, "libinfo-test.json"))
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	t.Run("Insert LibsBooks data into db", func(t *testing.T) {
		err := InsertLibsBooksRelationBulkFromJSON(testQuery, ctx, filepath.Join(cfg.DATA_PATH, "insert-books.json"), 127058)
		if err != nil {
			t.Fatal(err)
		}
	})

}
