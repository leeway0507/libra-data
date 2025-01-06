package book_scrap

import (
	"context"
	"fmt"
	"libraData/config"
	"libraData/db"
	"libraData/db/sqlc"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var cfg *config.EnvConfig = config.GetEnvConfig()

func TestBookScrap(t *testing.T) {
	ctx := context.Background()
	conn := db.ConnectPG(cfg.DATABASE_TEST_URL, ctx)

	testDataPath := filepath.Join(cfg.DATA_PATH, "test", "book_spec")
	bookScrapInstance := New(sqlc.New(conn), testDataPath)
	t.Run("separate scrap result", func(t *testing.T) {
		const targetPath = "/Users/yangwoolee/repo/libra-data/data/test/scrap/kyobo"
		err := bookScrapInstance.DistributeDataByIsbn(targetPath)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("update scrap result", func(t *testing.T) {
		err := bookScrapInstance.InsertToDB()
		if err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(testDataPath)
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
				fmt.Println("All files should have U char")
				continue
			}
			// rename for test
			err = os.Rename(filepath.Join(testDataPath, fileName),
				filepath.Join(testDataPath, fileName[1:]))
			if err != nil {
				t.Fatal(err)
			}

		}
	})

}
