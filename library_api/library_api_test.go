package library_api

import (
	"context"
	"libraData/config"
	"libraData/db"
	"libraData/db/sqlc"
	"path/filepath"
	"testing"
)

func TestLibraryAPI(t *testing.T) {
	cfg := config.GetEnvConfig()
	testDataPath := filepath.Join(cfg.DATA_PATH, "test", "library_api")

	libAPI := NewReq(111015, "2024-11-01", "2024-11-30", cfg.LIB_API_KEY, testDataPath)
	t.Run("test request => preprocess => save", func(t *testing.T) {
		// today := time.Now().Format(time.DateOnly)

		resp, err := libAPI.RequestBookData(1, 1)
		if err != nil {
			t.Fatal(err)
		}
		if len(resp.Docs) == 0 {
			t.Fatalf("Request Denied : %+v", resp)
		}
		docs, err := libAPI.Preprocess(resp)
		if err != nil {
			t.Fatal(err)
		}
		err = libAPI.Save(filepath.Join(testDataPath, "temp.json"), docs)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("test insert books libsbooks", func(t *testing.T) {

		ctx := context.Background()
		conn := db.ConnectPG(cfg.DATABASE_TEST_URL, ctx)
		defer conn.Close(ctx)

		db.InitTestTable(conn, ctx)

		testQuery := sqlc.New(conn)

		const libCode = 127058
		LibAPIDB := NewDB(testQuery, libCode, testDataPath)

		t.Run("books", func(t *testing.T) {
			err := LibAPIDB.InsertBooks(filepath.Join(testDataPath, "insert-books.json"))
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Run("libInfo", func(t *testing.T) {
			err := LibAPIDB.InsertLibInfo(filepath.Join(testDataPath, "libinfo-test.json"))
			if err != nil {
				t.Fatal(err)
			}
		})
		t.Run("libsbooks", func(t *testing.T) {
			err := LibAPIDB.InsertLibsBooks(filepath.Join(testDataPath, "insert-books.json"))
			if err != nil {
				t.Fatal(err)
			}
		})

	})

}
