package embedding

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"libraData/db/sqlc"
	"libraData/pb"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/protobuf/proto"
)

const ALREADY_UPDATED = "U"
const maxTokenSize = 8192

type EmbeddingData struct {
	Isbn      string
	Embedding []float32
}

type Req struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type Resp struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int32 `json:"prompt_tokens"`
		TotalTokens  int32 `json:"total_tokens"`
	} `json:"usage"`
}

type BatchResultResp struct {
	ID       string `json:"id"`
	CustomID string `json:"custom_id"`
	Response struct {
		StatusCode int32  `json:"status_code"`
		RequestID  string `json:"request_id"`
		Body       Resp
	} `json:"response"`
	Error any `json:"error"`
}

type BatchUploadReq struct {
	CustomId string `json:"custom_id"`
	Method   string `json:"method"`
	Url      string `json:"url"`
	Body     Req    `json:"body"`
}

type BatchExecReq struct {
	InputFileID      string `json:"input_file_id"`
	Endpoint         string `json:"endpoint"`
	CompletionWindow string `json:"completion_window"`
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

func (R *req) LoadBookDataFromJson(path string) []sqlc.ExtractBooksForEmbeddingRow {
	b := R.LoadFile(path)
	var books []sqlc.Book
	err := json.Unmarshal(b, &books)
	if err != nil {
		panic(err)
	}

	var bookForEmbedding []sqlc.ExtractBooksForEmbeddingRow
	for _, book := range books {
		bookForEmbedding = append(bookForEmbedding, sqlc.ExtractBooksForEmbeddingRow{
			Isbn:           book.Isbn,
			Title:          book.Title,
			Description:    book.Description,
			Toc:            book.Toc,
			Recommendation: book.Recommendation,
		})
	}
	return bookForEmbedding
}

func (R *req) LoadBookDataFromDB() []sqlc.ExtractBooksForEmbeddingRow {
	ctx := context.Background()
	data, err := R.query.ExtractBooksForEmbedding(ctx)
	if err != nil {
		panic(err)
	}
	return data
}

func (R *req) RequestEmbedding(data sqlc.ExtractBooksForEmbeddingRow) (*EmbeddingData, error) {
	reqBody := R.PrepareEmbeddingRequestBody(&data)

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

	var openAIresp Resp
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}
	json.Unmarshal(body, &openAIresp)

	return &EmbeddingData{
		Isbn:      data.Isbn.String,
		Embedding: openAIresp.Data[0].Embedding,
	}, nil
}

func (R *req) CreateBatchReqFile(rawData []sqlc.ExtractBooksForEmbeddingRow) ([]BatchUploadReq, error) {
	var batchList []BatchUploadReq
	for _, data := range rawData {
		batchList = append(batchList, BatchUploadReq{
			CustomId: data.Isbn.String,
			Method:   "POST",
			Url:      "/v1/embeddings",
			Body:     *R.PrepareEmbeddingRequestBody(&data),
		})
	}
	if len(batchList) == 0 {
		return nil, fmt.Errorf("no batch list")
	}
	return batchList, nil
}

func (R *req) SaveBatchReqFile(path string, batchReq []BatchUploadReq) {
	f, err := os.Create(path)
	if err != nil {
		log.Panic(err)
	}
	writer := bufio.NewWriter(f)

	for _, req := range batchReq {
		line, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}
		_, err = writer.WriteString(string(line) + "\n")
		if err != nil {
			panic(err)
		}
	}
	writer.Flush()
}

func (R *req) UploadBatchReqFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to open file: %v", err))
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	err = writer.WriteField("purpose", "batch")
	if err != nil {
		panic(err)
	}
	fileWriter, err := writer.CreateFormFile("file", path)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		panic(err)
	}
	err = writer.Close()
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/files", &body)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Bearer "+R.openAIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	uploadFileName := strings.Split(filepath.Base(path), ".")[0] + "_upload.json"
	f, err := os.Create(filepath.Join(filepath.Dir(path), uploadFileName))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Write(respBody)
}
func (R *req) ExecuteBatch(path string) {
	b := R.LoadFile(path)
	temp := make(map[string]any)

	err := json.Unmarshal(b, &temp)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal data: %v", err))
	}

	fileID, isExist := temp["id"]
	if !isExist {
		log.Printf("temp: %#+v\n", temp)
		panic(fmt.Sprintf("id is not exits %v", fileID))
	}

	strFileID, ok := fileID.(string)
	if !ok {
		panic(fmt.Sprintf("Value is not a string %v", fileID))
	}

	respBody := R.ExecuteBatchReq(strFileID)
	batchFileName := strings.Split(filepath.Base(path), "_upload.json")[0] + "_batch_start.json"
	f, err := os.Create(filepath.Join(filepath.Dir(path), batchFileName))
	if err != nil {
		panic(fmt.Sprintf("Failed to read response body: %v", err))
	}
	defer f.Close()
	f.Write(respBody)
}
func (R *req) ExecuteBatchReq(uploadFileId string) []byte {
	requestData := BatchExecReq{
		InputFileID:      uploadFileId,
		Endpoint:         "/v1/embeddings",
		CompletionWindow: "24h",
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal JSON: %v", err))
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/batches", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(fmt.Sprintf("Failed to create request: %v", err))
	}

	req.Header.Set("Authorization", "Bearer "+R.openAIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Failed to send request: %v", err))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Failed to read all: %v", err))
	}
	return b
}

func (R *req) GetBatchFileName(path string) map[string]string {
	b := R.LoadFile(path)
	temp := make(map[string]any)

	err := json.Unmarshal(b, &temp)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal: %v", err))
	}
	batchId, isExist := temp["id"]
	if !isExist {
		log.Printf("temp: %#+v\n", temp)
		panic(fmt.Sprintf("id is not exits %v", batchId))
	}
	strBatchId, ok := batchId.(string)
	if !ok {
		panic(fmt.Sprintf("Value is not a string %v", batchId))
	}

	body := R.Get(fmt.Sprintf("https://api.openai.com/v1/batches/%s", strBatchId))

	bodyBinary, err := io.ReadAll(body)
	if err != nil {
		panic(err)
	}
	respMap := make(map[string]any)

	err = json.Unmarshal(bodyBinary, &respMap)
	if err != nil {
		panic(err)
	}
	outputFileId, isExist := respMap["output_file_id"]
	if !isExist {
		log.Printf("respTemp: %#+v\n", respMap)
		panic(fmt.Sprintf("id is not exits %v", outputFileId))
	}
	strOutputFileId, _ := outputFileId.(string)

	errorFileId, isExist := respMap["error_file_id"]
	if !isExist {
		log.Printf("respTemp: %#+v\n", respMap)
		panic(fmt.Sprintf("id is not exits %v", errorFileId))
	}
	strErrorFileId, _ := errorFileId.(string)

	return map[string]string{
		"outputFileId": strOutputFileId,
		"errorFileId":  strErrorFileId,
	}
}

func (R *req) GetBatchRawData(path string) {
	result := R.GetBatchFileName(path)
	outputFileId, isExist := result["outputFileId"]

	if !isExist {
		log.Printf("x: %#+v\n", result)
		panic(fmt.Sprintf("id is not exits %v", outputFileId))
	}

	saveFile := strings.Split(filepath.Base(path), "_batch_start.json")[0] + "_batch_data.jsonl"
	savePath := filepath.Join(filepath.Dir(path), saveFile)

	body := R.Get(fmt.Sprintf("https://api.openai.com/v1/files/%s/content", outputFileId))

	out, err := os.Create(savePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to create file: %v", err))
	}
	defer out.Close()

	_, err = io.Copy(out, body)
	if err != nil {
		panic(fmt.Sprintf("Failed to write response to file: %v", err))
	}
	log.Printf("save: %#+v\n", savePath)
}

func (R *req) PreprocessBatchData(path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read file: %v", err))
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		var record BatchResultResp
		err := json.Unmarshal([]byte(line), &record)
		if err != nil {
			fmt.Printf("Failed to parse line: %s, error: %v\n", line, err)
			continue
		}
		R.SaveEmbeddingResp(&EmbeddingData{
			Isbn:      record.CustomID,
			Embedding: record.Response.Body.Data[0].Embedding,
		})
	}

	if err := scanner.Err(); err != nil {
		panic(fmt.Sprintf("Error reading file: %v", err))
	}

}

func (R *req) PrepareEmbeddingRequestBody(data *sqlc.ExtractBooksForEmbeddingRow) *Req {
	runes := []rune(data.Title.String +
		data.Description.String +
		data.Toc.String +
		data.Recommendation.String)

	reqBody := &Req{
		Input: string(runes[0:min(len(runes), maxTokenSize)]),
		Model: "text-embedding-3-small",
	}
	return reqBody
}

func (R *req) SaveEmbeddingResp(resp *EmbeddingData) error {
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
	b := R.LoadFile(filepath.Join(R.dataPath, isbn+".pb"))
	embeddingVector := &pb.EmbeddingVector{}
	err := proto.Unmarshal(b, embeddingVector)
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

func (R *req) Get(url string) io.ReadCloser {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to create request: %v", err))
	}
	req.Header.Set("Authorization", "Bearer "+R.openAIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(fmt.Sprintf("Failed to send request: %v", err))
	}

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Request failed with status: %s \n %s \n", resp.Status, url))
	}

	return resp.Body
}

func (R *req) LoadFile(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read file: %v", err))
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		panic(fmt.Sprintf("Failed to read all: %v", err))
	}
	return b
}
