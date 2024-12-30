package embedding

import (
	"context"
	"encoding/json"
	"io"
	"libraData/config"
	"libraData/db"
	"libraData/db/sqlc"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEmebedding(t *testing.T) {
	cfg := config.GetEnvConfig()
	ctx := context.Background()
	conn := db.ConnectPG(cfg.DATABASE_TEST_URL, ctx)
	defer conn.Close(ctx)

	testDataPath := filepath.Join(cfg.DATA_PATH, "test", "embedding")

	testQuery := sqlc.New(conn)
	req := NewReq(testQuery, cfg.OPEN_AI_API_KEY, testDataPath)

	t.Run("load bookdata from db", func(t *testing.T) {
		data := req.LoadBookDataFromDB()
		if len(data) == 0 {
			t.Fatal("data length is 0")
		}
	})
	t.Run("load bookdata from csv", func(t *testing.T) {
		data := req.LoadBookDataFromJson(filepath.Join(testDataPath, "embedding_example.json"))
		if len(data) == 0 {
			t.Fatal("data length is 0")
		}
		log.Printf("data: %#+v\n", data)
	})
	t.Run("Create batch Request", func(t *testing.T) {
		data := req.LoadBookDataFromJson(filepath.Join(testDataPath, "embedding_example.json"))
		batchReq, err := req.CreateBatchReqFile(data)
		if err != nil {
			t.Fatal(err)
		}
		if len(batchReq) == 0 {
			t.Fatal("data length is 0")
		}
		now := time.Now()
		filename := now.Format("2006-01-02_15-04-05") + ".jsonl"

		req.SaveBatchReqFile(filepath.Join(testDataPath, filename), batchReq)
		if _, err := os.Stat(filepath.Join(testDataPath, filename)); err != nil {
			if os.IsNotExist(err) {
				t.Fatal(err)
			}
		}
	})
	t.Run("Upload batch Request to openai server", func(t *testing.T) {
		req.UploadBatchReqFile(filepath.Join(testDataPath, "2024-12-30_12-27-00.jsonl"))
	})
	t.Run("Execute Batch Request", func(t *testing.T) {
		req.ExecuteBatch(filepath.Join(testDataPath, "2024-12-30_12-27-00_upload.json"))
	})
	t.Run("Download Batch Data", func(t *testing.T) {
		req.GetBatchRawData(filepath.Join(testDataPath, "2024-12-30_12-27-00_batch_start.json"))
	})
	t.Run("Preprocess Batch Data", func(t *testing.T) {
		req.PreprocessBatchData(filepath.Join(testDataPath, "2024-12-30_12-27-00_batch_data.jsonl"))
	})

	t.Run("save", func(t *testing.T) {
		file, err := os.Open(filepath.Join(testDataPath, "openai_resp.json"))
		if err != nil {
			t.Fatal(err)
		}
		b, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
		var openAIresp Resp

		json.Unmarshal(b, &openAIresp)

		vectors := &EmbeddingData{
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
