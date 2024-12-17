package embedding

import (
	"context"
	"encoding/json"
	"io"
	"libraData/db"
	"libraData/db/sqlc"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmebedding(t *testing.T) {

	ctx := context.Background()
	conn := db.ConnectPG(cfg.DATABASE_TEST_URL, ctx)
	cfg.DATA_PATH = filepath.Join(cfg.DATA_PATH, "test")
	defer conn.Close(ctx)

	// db.InitTestTable(conn, ctx)

	testQuery := sqlc.New(conn)
	t.Run("load", func(t *testing.T) {
		data := LoadDataForEmbedding(testQuery)
		if len(data) == 0 {
			t.Fatal("data length is 0")
		}
	})
	// t.Run("request embedding", func(t *testing.T) {
	// 	data := Load(testQuery)
	// 	RequestEmbedding(data[0])
	// })

	t.Run("save", func(t *testing.T) {
		file, err := os.Open(filepath.Join(cfg.DATA_PATH, "embedding", "openai_resp.json"))
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
		var openAIresp OpenAIEmbeddingResp

		json.Unmarshal(b, &openAIresp)

		vectors := &ResponseEmbedding{
			Isbn:      "9791138337526",
			Embedding: openAIresp.Data[0].Embedding,
		}

		if err = SaveEmbedding(vectors); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("load embedding", func(t *testing.T) {
		_, err := LoadEmbeddingData("9791138337526")
		if err != nil {
			t.Fatal(err)
		}
		// fmt.Printf("%+v", embedding.Embedding)
	})

	t.Run("update scrap result", func(t *testing.T) {
		embedding := filepath.Join(cfg.DATA_PATH, "embedding")
		err := UpdateEmbeddingFromPB(testQuery, ctx)
		if err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(filepath.Join(cfg.DATA_PATH, "embedding"))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			fileName := entry.Name()
			_, isPB := strings.CutSuffix(fileName, ".pb")
			if !isPB {
				continue
			}
			if fileName[:1] != "U" {
				t.Fatal("All files should have U char")
				continue
			}
			// rename for test
			err = os.Rename(filepath.Join(embedding, fileName),
				filepath.Join(embedding, fileName[1:]))
			if err != nil {
				t.Fatal(err)
			}

		}
	})
}
