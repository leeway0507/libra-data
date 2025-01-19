package scrap

import (
	"encoding/json"
	"fmt"
	"io"
	"libraData/pkg/utils"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type DaumDocument struct {
	Authors     []string `json:"authors"`
	Contents    string   `json:"contents"`
	Datetime    string   `json:"datetime"`
	ISBN        string   `json:"isbn"`
	Price       int      `json:"price"`
	Publisher   string   `json:"publisher"`
	SalePrice   int      `json:"sale_price"`
	Status      string   `json:"status"`
	Thumbnail   string   `json:"thumbnail"`
	Title       string   `json:"title"`
	Translators []string `json:"translators"`
	URL         string   `json:"url"`
}

type DaumMeta struct {
	IsEnd         bool `json:"is_end"`
	PageableCount int  `json:"pageable_count"`
	TotalCount    int  `json:"total_count"`
}

type DaumResponse struct {
	Documents []DaumDocument `json:"documents"`
	Meta      DaumMeta       `json:"meta"`
}

var (
	DaumKey = os.Getenv("DAUM_KEY")
)

func RequestDaumAll(query []string, path string, workers int) {
	tasks := make(chan string, len(query))

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for query := range tasks {
				RequestDaum(query, path)
				time.Sleep(time.Duration(rand.Intn(2)) * time.Second)

			}

		}(i)
	}
	for _, q := range query {
		tasks <- q
	}
	close(tasks)
	wg.Wait()
}

func RequestDaum(query, dir string) {
	if DaumKey == "" {
		log.Fatal("failed to load daum key from OS ENV")
	}
	// check the file exists
	path := filepath.Join(dir, query+".json")
	uPath := filepath.Join(dir, "U"+query+".json")
	if utils.CheckFileExist(path) || utils.CheckFileExist(uPath) {
		log.Printf("query '%s' exists", query)
		return
	}

	// request daum
	baseURL := "https://dapi.kakao.com/v3/search/book"
	rawURL, err := url.Parse(baseURL)
	utils.HandleErr(err, "Parsing URL")

	queryParams := rawURL.Query()
	queryParams.Set("query", query)
	rawURL.RawQuery = queryParams.Encode()

	req, err := http.NewRequest("GET", rawURL.String(), nil)
	utils.HandleErr(err, "Creating request")
	req.Header.Add("Authorization", fmt.Sprintf("KakaoAK %s", DaumKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	utils.HandleErr(err, "Executing request")
	defer resp.Body.Close()

	// save daum response
	body, err := io.ReadAll(resp.Body)
	utils.HandleErr(err, "Reading response body")

	if resp.StatusCode != 200 {
		utils.HandleErr(fmt.Errorf("daum_status %v", resp.StatusCode), fmt.Sprintf("%v", string(body)))
		return
	}

	var daumBody DaumResponse
	err = json.Unmarshal(body, &daumBody)
	utils.HandleErr(err, "Unmarshalling response")

	if len(daumBody.Documents) == 0 {
		log.Printf("Daum_ No items found for query: %s status %v \n ", query, resp.StatusCode)
		file, err := os.Create(path)
		utils.HandleErr(err, "Creating file")
		// save empty result
		b, err := json.Marshal(&BookResp{
			Isbn:   query,
			Source: "daum",
		})
		utils.HandleErr(err, "marshal json")
		file.Write(b)
		return
	}
	daum := daumBody.Documents[0]

	file, err := os.Create(path)
	utils.HandleErr(err, "Creating file")
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)

	result := &BookResp{
		Isbn:        query,
		Title:       daum.Title,
		Author:      strings.Join(daum.Authors, " "),
		ImageUrl:    daum.Thumbnail,
		Description: daum.Contents,
		Source:      "daum",
		Url:         daum.URL,
	}
	utils.HandleErr(err, "Marshalling item")

	err = encoder.Encode(result)
	utils.HandleErr(err, "Writing to file")
}
