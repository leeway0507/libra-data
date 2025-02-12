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
	"sync"
	"time"
)

type NaverDocument struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Image       string `json:"image"`
	Author      string `json:"author"`
	Discount    string `json:"discount"`
	Publisher   string `json:"publisher"`
	PubDate     string `json:"pubdate"`
	ISBN        string `json:"isbn"`
	Description string `json:"description"`
}

type NaverResponse struct {
	LastBuildDate string          `json:"lastBuildDate"`
	Total         int             `json:"total"`
	Start         int             `json:"start"`
	Display       int             `json:"display"`
	Items         []NaverDocument `json:"items"`
}

var (
	clientId     = os.Getenv("NAVER_ID")
	clientSecret = os.Getenv("NAVER_SECRET")
)

func RequestNaverAll(query []string, path string, workers int) {
	tasks := make(chan string, len(query))

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for query := range tasks {
				RequestNaver(query, path)
				time.Sleep(time.Duration(1+rand.Float32()) * time.Second)
			}

		}(i)
	}
	for _, q := range query {
		tasks <- q
	}
	close(tasks)
	wg.Wait()
}

func RequestNaver(query, dir string) {
	if clientId == "" || clientSecret == "" {
		log.Fatal("failed to load naver client id or clientSecret from OS ENV")
	}
	// check the file exists
	path := filepath.Join(dir, query+".json")
	uPath := filepath.Join(dir, "U"+query+".json")
	if utils.CheckFileExist(path) || utils.CheckFileExist(uPath) {
		log.Printf("naver query '%s' exists", query)
		//temp
		os.Remove(uPath)

		// return
	}

	// request naver
	baseURL := "https://openapi.naver.com/v1/search/book.json"
	rawURL, err := url.Parse(baseURL)
	utils.HandleErr(err, "Parsing URL")

	queryParams := rawURL.Query()
	queryParams.Set("query", query)
	rawURL.RawQuery = queryParams.Encode()

	req, err := http.NewRequest("GET", rawURL.String(), nil)
	utils.HandleErr(err, "Creating request")
	req.Header.Add("X-Naver-Client-Id", clientId)
	req.Header.Add("X-Naver-Client-Secret", clientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	utils.HandleErr(err, "Executing request")
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	utils.HandleErr(err, "Reading response body")

	if resp.StatusCode != 200 {
		utils.HandleErr(fmt.Errorf("naver_status %v", resp.StatusCode), fmt.Sprintf("%v", string(body)))
		return
	}

	// save naver response
	var naverBody NaverResponse
	err = json.Unmarshal(body, &naverBody)
	utils.HandleErr(err, "Unmarshalling response")

	if len(naverBody.Items) == 0 {
		log.Printf("Naver_ No items found for query: %s status %v \n", query, resp.StatusCode)
		file, err := os.Create(path)
		utils.HandleErr(err, "Creating file")
		// save empty result
		b, err := json.Marshal(&BookResp{
			Isbn:   query,
			Source: "naver",
		})
		utils.HandleErr(err, "marshal json")
		file.Write(b)
		return
	}

	naver := naverBody.Items[0]
	utils.HandleErr(err, "Marshalling item")

	file, err := os.Create(path)
	utils.HandleErr(err, "Creating file")
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	result := &BookResp{
		Isbn:        query,
		Title:       naver.Title,
		Author:      naver.Author,
		ImageUrl:    naver.Image,
		Description: naver.Description,
		Source:      "naver",
		Url:         naver.Link,
	}
	err = encoder.Encode(result)
	utils.HandleErr(err, "encode result")

}
