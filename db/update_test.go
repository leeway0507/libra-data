package db

import (
	"context"
	"libraData/config"
	sqlc "libraData/db/sqlc"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Update(t *testing.T) {
	config := config.GetEnvConfig()

	ctx := context.Background()
	conn := ConnectPG(config.DATABASE_TEST_URL, ctx)
	defer conn.Close(ctx)

	err := InitTestTable(conn, ctx)
	if err != nil {
		t.Fatal(err)
	}
	testQuery := sqlc.New(conn)
	defaultPath := "/Users/yangwoolee/repo/libra-data/data/test"

	t.Run("update scrap result", func(t *testing.T) {
		isbnPath := filepath.Join(defaultPath, "isbn")
		err := UpdateScrapResultFromJSON(testQuery, ctx, isbnPath)
		if err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(isbnPath)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			// check file name changed from 1234 to U1234
			fileName := entry.Name()
			_, isJson := strings.CutSuffix(fileName, ".json")
			if !isJson {
				continue
			}
			if fileName[:1] != ALREADY_UPDATED {
				t.Fatal("All files should have U char")
				continue
			}
			// rename for test
			err = os.Rename(filepath.Join(isbnPath, fileName),
				filepath.Join(isbnPath, fileName[1:]))
			if err != nil {
				t.Fatal(err)
			}

		}
	})
	t.Run("separate scrap result", func(t *testing.T) {
		const targetPath = "/Users/yangwoolee/repo/libra-data/data/test/scrap/kyobo"
		const savePath = "/Users/yangwoolee/repo/libra-data/data/test/isbn"
		err := SeparateScrapResultByEachISBN(targetPath, savePath)
		if err != nil {
			t.Fatal(err)
		}
	})
}
