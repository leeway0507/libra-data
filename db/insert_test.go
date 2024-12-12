package db

import (
	"context"
	"libraData/config"
	sqlc "libraData/db/sqlc"

	"path/filepath"
	"testing"
)

func Test_Insert(t *testing.T) {
	config := config.GetEnvConfig()

	ctx := context.Background()
	conn := connectPG(config.DATABASE_TEST_URL, ctx)
	defer conn.Close(ctx)
	// err := CreateTestTable(conn, ctx)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer func() {
	// 	err := DropTestTable(conn, ctx)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// }()

	testQuery := sqlc.New(conn)
	defaultPath := "/Users/yangwoolee/repo/libra-data/data/test"

	t.Run("Insert books into db", func(t *testing.T) {
		err := InsertLibBookBulkFromJSON(testQuery, ctx, filepath.Join(defaultPath, "insert-books.json"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := testQuery.DeleteAllBooksRowForTest(ctx); err != nil {
			t.Fatal(err)
		}

	})
	t.Run("Insert libinfos into db", func(t *testing.T) {

		err := InsertLibInfoBulkFromJSON(testQuery, ctx, filepath.Join(defaultPath, "libinfo-test-updated.json"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := testQuery.DeleteAllLibariesForTest(ctx); err != nil {
			t.Fatal(err)
		}
	})

}
