package library_api

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"libraData/config"
	"libraData/db"
	"libraData/db/sqlc"
	"libraData/utils"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type searchResponse struct {
	XMLName xml.Name `xml:"response"`
	Header  Header   `xml:"header"`
	Total   int      `xml:"total"`
	Record  Record   `xml:"recode"`
}

type Header struct {
	ResultMsg  string `xml:"resultMsg"`
	ResultCode string `xml:"resultCode"`
}

type Record struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

type tocResponse struct {
	XMLName xml.Name `xml:"response"`
	TOC     TOC      `xml:"toc"`
}

type TOC struct {
	Content string `xml:",innerxml"`
}

var (
	searchUrl      = "http://apis.data.go.kr/9720000/searchservice/detail"
	tocUrl         = "http://apis.data.go.kr/9720000/detailinfoservice/toc"
	userAgent      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	cfg            = config.GetEnvConfig()
	UPDATED        = "U"
	NOT_EXIST_ISBN = "N"
)

type congress struct {
	query    *sqlc.Queries
	dataPath string
}

func RequestCongress(isbns []string, path string, workers int) {
	ctx := context.Background()
	pool := db.ConnectPGPool(cfg.DATABASE_URL, ctx)
	tasks := make(chan string, len(isbns))

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			conn, err := pool.Acquire(ctx)
			if err != nil {
				panic(err)
			}
			defer wg.Done()
			defer conn.Release()
			congress := NewCongress(sqlc.New(conn), path)
			for isbn := range tasks {
				err := congress.RequestDetail(isbn)
				if err != nil {
					log.Println(err)
					continue
				}
				congress.ReqeustToc(isbn)
				time.Sleep(time.Duration(rand.Intn(1)) * time.Second)
			}

		}(i)
	}
	for _, i := range isbns {
		tasks <- i
	}
	close(tasks)
	wg.Wait()
}

func NewCongress(query *sqlc.Queries, dataPath string) *congress {
	return &congress{
		query,
		dataPath,
	}
}

func (c *congress) GetTargetFromDB() []string {
	ctx := context.Background()
	result, err := c.query.GetBooksWithoutToc(ctx)
	utils.HandleErr(err, "getTargetFromDB")
	isbns := []string{}
	for _, v := range result {
		isbns = append(isbns, v.String)
	}
	return isbns
}

func (c *congress) RequestDetail(isbn string) error {
	if utils.CheckFileExist(filepath.Join(c.dataPath, "detail", UPDATED+isbn+".json")) {
		return fmt.Errorf("%s alreay exists", isbn)
	}

	url, err := url.Parse(searchUrl)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	queryParam := url.Query()
	queryParam.Add("serviceKey", cfg.PUBLIC_DATA_PORTAL_KEY)
	queryParam.Add("pageno", "1")
	queryParam.Add("displaylines", "10")
	queryParam.Add("dbname", "일반도서")
	queryParam.Add("search", fmt.Sprintf("ISBN,%s", isbn))

	url.RawQuery = queryParam.Encode()
	req, _ := http.NewRequest("GET", url.String(), nil)
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	f, err := os.Create(filepath.Join(c.dataPath, "detail", isbn+".json"))
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	c.SaveRespAsJson(b, f)
	return nil
}

func (c *congress) SaveRespAsJson(data []byte, buf *os.File) {
	var xmlFile searchResponse
	err := xml.Unmarshal(data, &xmlFile)
	if err != nil {
		log.Println(string(data))
		return
	}

	result := map[string]string{}
	for _, item := range xmlFile.Record.Items {
		name := item.Name
		value := item.Value
		result[name] = value
	}
	jsonData, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	_, err = buf.Write(jsonData)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *congress) ReqeustToc(isbn string) {
	b, err := os.ReadFile(filepath.Join(c.dataPath, "detail", isbn+".json"))
	if err != nil {
		panic(err)
	}
	var detail map[string]string
	json.Unmarshal(b, &detail)

	if len(detail) == 0 {
		log.Printf("%s len 0 ", isbn)
		f, err := os.Create(filepath.Join(c.dataPath, "toc", NOT_EXIST_ISBN+isbn+".json"))
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		f.Write([]byte(""))

		err = os.Rename(filepath.Join(c.dataPath, "detail", isbn+".json"),
			filepath.Join(c.dataPath, "detail", UPDATED+isbn+".json"))
		if err != nil {
			log.Println(err)
		}
		c.InsertNone(isbn)

		return
	}

	var controlno string
	for key, value := range detail {
		if key == "제어번호" {
			controlno = value
		}
	}

	url, err := url.Parse(tocUrl)
	if err != nil {
		panic(err)
	}
	queryParam := url.Query()
	queryParam.Add("serviceKey", cfg.PUBLIC_DATA_PORTAL_KEY)
	queryParam.Add("controlno", controlno)

	url.RawQuery = queryParam.Encode()
	req, _ := http.NewRequest("GET", url.String(), nil)
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	utils.HandleErr(err, "req")

	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	utils.HandleErr(err, "readAll")

	c.SaveTocAsJson(b, isbn)
}

func (c *congress) SaveTocAsJson(data []byte, isbn string) {
	var xmlFile tocResponse
	err := xml.Unmarshal(data, &xmlFile)
	if err != nil {
		log.Println(string(data))
		return
	}

	htmlDecoded := html.UnescapeString(xmlFile.TOC.Content)
	htmlDecoded = strings.ReplaceAll(htmlDecoded, "<BR>", "\n")
	htmlDecoded = strings.ReplaceAll(htmlDecoded, "<p>", "")

	toc, err := json.Marshal(sqlc.UpdateTocParams{
		Toc:  pgtype.Text{String: htmlDecoded, Valid: true},
		Isbn: pgtype.Text{String: isbn, Valid: true},
	})
	if err != nil {
		panic(err)
	}

	// rename for test
	f, err := os.Create(filepath.Join(c.dataPath, "toc", isbn+".json"))
	utils.HandleErr(err, "create")

	_, err = f.Write(toc)
	if err != nil {
		log.Fatal(err)
	}

	err = os.Rename(filepath.Join(c.dataPath, "detail", isbn+".json"),
		filepath.Join(c.dataPath, "detail", UPDATED+isbn+".json"))
	utils.HandleErr(err, "rename")
}

func (c *congress) InsertAll() error {
	isbnPath := filepath.Join(c.dataPath, "toc")
	entries, err := os.ReadDir(isbnPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fileName := entry.Name()
		err := c.Insert(fileName)
		if err != nil {
			utils.HandleErr(err, "insertToDB")
		}

	}
	return nil
}

func (c *congress) Insert(fileName string) error {
	isbnPath := filepath.Join(c.dataPath, "toc")
	_, isJson := strings.CutSuffix(fileName, ".json")
	if !isJson {
		log.Printf("found non json file %s", fileName)
		return nil
	}
	if slices.Contains([]string{UPDATED, NOT_EXIST_ISBN}, fileName[:1]) {
		// log.Printf("%v already updated \n", fileName)
		return nil
	}
	b, err := os.ReadFile(filepath.Join(isbnPath, fileName))
	if err != nil {
		return err
	}
	var jsonArr sqlc.UpdateTocParams
	if err = json.Unmarshal(b, &jsonArr); err != nil {
		return err
	}
	ctx := context.Background()

	if err = c.query.UpdateToc(ctx, jsonArr); err != nil {
		return err
	}

	err = os.Rename(filepath.Join(isbnPath, fileName),
		filepath.Join(isbnPath, UPDATED+fileName))
	if err != nil {
		return err
	}
	return nil
}

func (c *congress) InsertNoneAll() error {
	isbnPath := filepath.Join(c.dataPath, "toc")
	entries, err := os.ReadDir(isbnPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fileName := entry.Name()
		err := c.InsertNone(fileName)
		if err != nil {
			log.Printf("err: %#+v\n", err)
		}

	}

	return nil
}
func (c *congress) InsertNone(fileName string) error {
	ctx := context.Background()
	if !slices.Contains([]string{NOT_EXIST_ISBN}, fileName[:1]) {
		return nil
	}
	fileName, _ = strings.CutSuffix(fileName, ".json")
	c.query.UpdateToc(ctx, sqlc.UpdateTocParams{
		Toc:  pgtype.Text{String: "", Valid: true},
		Isbn: pgtype.Text{String: fileName[1:], Valid: true},
	})
	return nil
}
