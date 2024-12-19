package preprocess

import (
	"fmt"
	"libraData/config"
	"os"
	"path/filepath"
	"testing"
)

func TestProprocess(t *testing.T) {
	config.SetTestEnvConfig(cfg)

	t.Run("check preprocess status", func(t *testing.T) {
		scrapDate := "2024-12-01"

		// remove existing test file
		_, err := os.Open(filepath.Join(cfg.DATA_PATH, "library", "가락몰도서관", scrapDate+".pb"))
		if err == nil {
			fmt.Println("pb file exists. removing the file...")
			os.Remove(filepath.Join(cfg.DATA_PATH, "library", "가락몰도서관", scrapDate+".pb"))
		}

		libEntries := LoadLibScraperFolder()
		if len(libEntries) == 0 {
			t.Fatal("no lib folders")
		}

		for _, libEntry := range libEntries {
			ep := NewExcelToProto(libEntry, scrapDate)
			isPreprocessed := ep.GetPreprocessStatus()

			if isPreprocessed {
				continue
			}

			err := ep.Preprocess()
			if err != nil {
				t.Fatal(err)
			}
			books, err := ep.LoadBooksFromPB()
			if err != nil {
				t.Fatal(err)
			}
			if len(books.Books) == 0 {
				t.Fatal("can not load book file")
			}

		}
	})
	t.Run("load Book Rows", func(t *testing.T) {
	})

}
