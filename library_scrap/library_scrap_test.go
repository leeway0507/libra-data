package library_scrap

import (
	"fmt"
	"libraData/config"
	"os"
	"path/filepath"
	"testing"
)

func TestProprocess(t *testing.T) {
	cfg := config.GetEnvConfig()

	t.Run("check preprocess status", func(t *testing.T) {
		scrapDate := "2024-12-01"
		testDataPath := filepath.Join(cfg.DATA_PATH, "test", "library")

		// remove existing test file
		_, err := os.Open(filepath.Join(testDataPath, "가락몰도서관", scrapDate+".pb"))
		if err == nil {
			fmt.Println("pb file exists. removing the file...")
			os.Remove(filepath.Join(testDataPath, "가락몰도서관", scrapDate+".pb"))
		}

		libEntries := LoadLibScrapFolders(testDataPath)
		if len(libEntries) == 0 {
			t.Fatal("no lib folders")
		}

		for _, libEntry := range libEntries {
			ep := NewExcelToProto(libEntry, scrapDate, testDataPath)
			isPreprocessed := ep.GetPreprocessStatus()

			if isPreprocessed {
				continue
			}

			err := ep.Preprocess()
			if err != nil {
				t.Fatal(err)
			}

		}
	})
	t.Run("load Book Rows", func(t *testing.T) {
	})

}
