package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"libraData/db/sqlc"
	"libraData/pb"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/protobuf/proto"
)

const ALREADY_UPDATED = "U"
const maxTokenSize = 10

type Vector struct {
	Isbn   string
	Vector []float32
}

type RequestEmbeddingBody struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type OpenAIEmbeddingResp struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		Prompt_tokens int
		Potal_tokens  int
	} `json:"usage"`
}

type ResponseEmbedding struct {
	Isbn      string
	Embedding []float32
}

type req struct {
	query     *sqlc.Queries
	openAIKey string
	dataPath  string
}

func NewReq(query *sqlc.Queries, openAIKey string, dataPath string) *req {
	return &req{
		query,
		openAIKey,
		dataPath,
	}
}

type DB struct {
	Isbn      string
	Embedding []float32
}

func (R *req) LoadBookData() []sqlc.ExtractBooksForEmbeddingRow {
	ctx := context.Background()
	data, err := R.query.ExtractBooksForEmbedding(ctx)
	if err != nil {
		panic(err)
	}
	return data
}

func (R *req) RequestEmbedding(data sqlc.ExtractBooksForEmbeddingRow) (*ResponseEmbedding, error) {
	runes := []rune(data.Title.String +
		data.Description.String +
		data.Toc.String +
		data.Recommendation.String)

	reqBody := &RequestEmbeddingBody{
		Input: string(runes[0:min(len(runes), maxTokenSize)]),
		Model: "text-embedding-3-small",
	}
	reqBodyByte, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := "https://api.openai.com/v1/embeddings"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBodyByte))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+R.openAIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var openAIresp OpenAIEmbeddingResp
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}
	json.Unmarshal(body, &openAIresp)

	return &ResponseEmbedding{
		Isbn:      data.Isbn.String,
		Embedding: openAIresp.Data[0].Embedding,
	}, nil
}

func (R *req) PrepareEmbeddingRequestBody(data *sqlc.ExtractBooksForEmbeddingRow) (*[]byte, error) {
	runes := []rune(data.Title.String +
		data.Description.String +
		data.Toc.String +
		data.Recommendation.String)

	reqBody := &RequestEmbeddingBody{
		Input: string(runes[0:min(len(runes), maxTokenSize)]),
		Model: "text-embedding-3-small",
	}
	reqBodyByte, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err

	}
	return &reqBodyByte, nil
}

func (R *req) SaveEmbeddingResp(resp *ResponseEmbedding) error {
	embeddingpb := &pb.EmbeddingVector{
		Embedding: resp.Embedding,
		Isbn:      resp.Isbn,
	}

	file, err := os.Create(filepath.Join(R.dataPath, resp.Isbn+".pb"))
	if err != nil {
		fmt.Println("file open error", err)
		return err
	}
	defer file.Close()
	b, err := proto.Marshal(embeddingpb)
	if err != nil {
		return err
	}
	file.Write(b)
	return nil
}

func (R *req) LoadEmbeddingData(isbn string) (*pb.EmbeddingVector, error) {
	file, err := os.Open(filepath.Join(R.dataPath, isbn+".pb"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	embeddingVector := &pb.EmbeddingVector{} // 포인터 초기화
	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(b, embeddingVector)
	if err != nil {
		return nil, err
	}

	return embeddingVector, nil
}

func (R *req) InsertToDB() error {

	embeddingPath := filepath.Join(R.dataPath)
	entries, err := os.ReadDir(embeddingPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fileName := entry.Name()
		isbn, isPB := strings.CutSuffix(fileName, ".pb")
		if !isPB {
			log.Println("found non PB file", fileName)
			continue
		}
		if isbn[:1] == ALREADY_UPDATED {
			log.Printf("%v is already upldated", fileName)
			continue
		}

		data, err := R.LoadEmbeddingData(isbn)
		if err != nil {
			log.Printf("loadembedding Error : %v \n", isbn)
			log.Println("move onto next data...")
			continue
		}
		embeddingArgs := sqlc.InsertEmbeddingsParams{
			Isbn:      isbn,
			Embedding: pgvector.NewVector(data.Embedding),
		}

		ctx := context.Background()
		err = R.query.InsertEmbeddings(ctx, embeddingArgs)
		if err != nil {
			return fmt.Errorf("insertEmbeddings %v \n \n error: %v", isbn, err)
		}
		vectorStatusArgs := sqlc.UpdateVectorSearchStatusParams{
			Isbn:         pgtype.Text{String: isbn, Valid: true},
			VectorSearch: pgtype.Bool{Bool: true, Valid: true},
		}

		err = R.query.UpdateVectorSearchStatus(ctx, vectorStatusArgs)
		if err != nil {
			return fmt.Errorf("updateVectorSearchStatus : %v \n error: %v", isbn, err)

		}
		err = os.Rename(filepath.Join(embeddingPath, fileName),
			filepath.Join(embeddingPath, "U"+fileName))
		if err != nil {
			return fmt.Errorf("rename : %v \n error: %v", isbn, err)

		}
	}
	return nil
}
