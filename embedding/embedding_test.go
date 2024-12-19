package embedding

import (
	"context"
	"encoding/json"
	"io"
	"libraData/config"
	"libraData/db"
	"libraData/db/sqlc"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmebedding(t *testing.T) {
	cfg := config.GetEnvConfig()
	ctx := context.Background()
	conn := db.ConnectPG(cfg.DATABASE_TEST_URL, ctx)
	defer conn.Close(ctx)

	testDataPath := filepath.Join(cfg.DATA_PATH, "test", "embedding")

	testQuery := sqlc.New(conn)
	req := NewReq(testQuery, cfg.OPEN_AI_API_KEY, testDataPath)
	t.Run("load", func(t *testing.T) {
		data := req.LoadBookData()
		if len(data) == 0 {
			t.Fatal("data length is 0")
		}
	})

	// t.Run("request embedding", func(t *testing.T) {
	// 	data := Load(testQuery)
	// 	RequestEmbedding(data[0])
	// })

	t.Run("save", func(t *testing.T) {
		file, err := os.Open(filepath.Join(testDataPath, "openai_resp.json"))
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

		if err = req.SaveEmbeddingResp(vectors); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("load embedding", func(t *testing.T) {
		_, err := req.LoadEmbeddingData("9791138337526")
		if err != nil {
			t.Fatal(err)
		}
		// fmt.Printf("%+v", embedding.Embedding)
	})

	t.Run("update scrap result", func(t *testing.T) {
		conn.Exec(ctx, "DELETE FROM bookembedding")
		embedding := filepath.Join(testDataPath)
		err := req.InsertToDB()
		if err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(filepath.Join(testDataPath))
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
