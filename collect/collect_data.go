package collect

import (
	"encoding/json"
	"fmt"
	"io"
	"libraData/config"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// http://data4library.kr/api/itemSrch?authKey=[발급받은키]&libCode=[도서관코드]&startDt=[검색시작일자]&
// endDt=[검색종료일자]&pageNo=1&pageSize=10

var cfg config.EnvConfig = *config.GetEnvConfig()

func GetBookItemsAll(libCode int, startDate string, endDate string) error {
	const pageSize = 500

	initResp, err := GetBookItems(libCode, startDate, endDate, 1, pageSize)
	if err != nil {
		return err
	}
	totalPage := ceilDiv(initResp.NumFound, pageSize)
	fmt.Printf("total book count : %v", initResp.NumFound)
	fmt.Printf("Planned Request Page : %v", totalPage)

	//create Temp folder name
	currPath, err := os.Getwd()
	if err != nil {
		return err
	}

	folderName := strings.Join([]string{strconv.Itoa(libCode), startDate, endDate, strconv.Itoa(pageSize), strconv.Itoa(totalPage)}, "-")
	folderPath := filepath.Join(currPath, "data/temp", folderName)

	// make folders
	if _, err = os.ReadDir(folderPath); err != nil {
		err = os.MkdirAll(folderPath, 0750)
		if err != nil {
			return err
		}
	}

	bar := progressbar.Default(int64(totalPage), "PageCollection")

	for idx := range totalPage {
		currPage := idx + 1
		fileName := filepath.Join(folderPath, strconv.Itoa(currPage)+".json")

		if _, err := os.Stat(fileName); err != nil {
			if os.IsNotExist(err) {
				resp, err := GetBookItems(libCode, startDate, endDate, currPage, pageSize)
				if err != nil {
					return err
				}
				docs, err := PreprocessBookItems(resp)
				if err != nil {
					return err
				}
				err = SaveBookItemsAsJson(fileName, docs)
				if err != nil {
					return err
				}
			}
		}
		bar.Add(1)
	}
	return nil
}

func GetBookItems(libCode int, startDate string, endDate string, pageNo int, pageSize int) (*BookItemsResponse, error) {
	url, err := url.Parse("http://data4library.kr/api/itemSrch")
	if err != nil {
		return nil, err
	}
	queryParam := url.Query()
	queryParam.Set("authKey", cfg.LIB_API_KEY)
	queryParam.Set("libCode", strconv.Itoa(libCode))
	queryParam.Set("startDt", startDate)
	queryParam.Set("endDt", endDate)
	queryParam.Set("format", "json")
	queryParam.Set("pageNo", strconv.Itoa(pageNo))
	queryParam.Set("pageSize", strconv.Itoa(pageSize))

	url.RawQuery = queryParam.Encode()
	resp, err := http.Get(url.String())
	if err != nil {
		return nil, err
	}
	bodyRaw := resp.Body
	defer bodyRaw.Close()

	b, err := io.ReadAll(bodyRaw)
	if err != nil {
		return nil, err
	}
	var bookItemsResp struct {
		Response BookItemsResponse `json:"response"`
	}
	err = json.Unmarshal(b, &bookItemsResp)
	if err != nil {
		return nil, err
	}
	return &bookItemsResp.Response, nil

}
func PreprocessBookItems(resp *BookItemsResponse) (*[]BookItemsDoc, error) {
	var docs []BookItemsDoc
	for _, doc := range resp.Docs {
		docs = append(docs, doc.Doc)
	}
	if docs == nil {
		return nil, fmt.Errorf("docs is empty")
	}
	return &docs, nil
}
func SaveBookItemsAsJson(jsonPath string, bookItems *[]BookItemsDoc) error {

	jsonFile, err := os.Create(jsonPath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	err = encoder.Encode(bookItems)
	if err != nil {
		return err
	}
	return nil
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}
