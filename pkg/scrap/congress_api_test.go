package scrap

import (
	"context"
	"libraData/pkg/db"
	"libraData/pkg/db/sqlc"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var (
	testDataPath = filepath.Join(cfg.DATA_PATH, "test", "library_api", "congress")
	isbn         = "9791168330702"
)

func TestCongress(t *testing.T) {
	conn := db.ConnectPG(cfg.DATABASE_URL, context.Background())
	query := sqlc.New(conn)
	congress := NewCongress(query, testDataPath)
	t.Run("request detail", func(t *testing.T) {
		congress.RequestBookDetail(isbn)

		if _, err := os.Stat(filepath.Join(testDataPath, "detail", isbn+".json")); err != nil {
			if os.IsNotExist(err) {
				t.Fatal(err)
			}
		}
	})
	t.Run("request isbn", func(t *testing.T) {
		congress.ReqeustToc(isbn)
		if _, err := os.Stat(filepath.Join(testDataPath, "toc", isbn+".json")); err != nil {
			if os.IsNotExist(err) {
				t.Fatal(err)
			}
		}
	})
	t.Run("get isbns", func(t *testing.T) {
		isbns := congress.ExtractBookISBNs()
		if len(isbns) == 0 {
			t.Fatal("isbn is 0")
		}
		log.Printf("len(isbns): %#+v\n", len(isbns))
	})
	t.Run("request detail failed", func(t *testing.T) {
		nonExistIsbn := "9788909207683"
		congress.RequestBookDetail(nonExistIsbn)
	})
	t.Run("request isbn failed", func(t *testing.T) {
		nonExistIsbn := "9788909207683"
		congress.ReqeustToc(nonExistIsbn)
	})

}
