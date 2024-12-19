package library_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"libraData/db/sqlc"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/schollz/progressbar/v3"
)

type Req struct {
	libCode   int
	startDate string
	endDate   string
	dataPath  string
	libAPIKey string
	PageSize  int
}

func NewReq(libCode int, startDate string, endDate string, libAPIKey string, dataPath string) *Req {
	return &Req{libCode, startDate, endDate, dataPath, libAPIKey, 500}
}

func (L *Req) RequestAndSave() error {

	// get max page size
	initResp, err := L.RequestBookData(1, 1)
	if err != nil {
		return err
	}
	totalPage := ceilDiv(initResp.NumFound, L.PageSize)
	fmt.Printf("total book count : %v \t", initResp.NumFound)
	fmt.Printf("Planned Request Page : %v \n", totalPage)

	// make folders
	folderName := strings.Join([]string{L.startDate, L.endDate, strconv.Itoa(L.PageSize), strconv.Itoa(totalPage)}, "-")
	folderPath := filepath.Join(L.dataPath, strconv.Itoa(L.libCode), folderName)
	if _, err = os.ReadDir(folderPath); err != nil {
		err = os.MkdirAll(folderPath, 0750)
		if err != nil {
			return err
		}
	}

	bar := progressbar.Default(int64(totalPage), "PageCollection")

	// collect and save data from library API
	for idx := range totalPage {
		currPage := idx + 1
		fileName := filepath.Join(folderPath, strconv.Itoa(currPage)+".json")

		if _, err := os.Stat(fileName); err != nil {
			if os.IsNotExist(err) {
				resp, err := L.RequestBookData(currPage, L.PageSize)
				if err != nil {
					fmt.Printf("Error : RequestBookData %v \n", err)
					continue
				}
				docs, err := L.Preprocess(resp)
				if err != nil {
					return fmt.Errorf("PreprocessBookItems : %v", err)
				}
				err = L.Save(fileName, docs)
				if err != nil {
					return fmt.Errorf("SaveBookItemsAsJson : %v", err)
				}
			} else {
				return fmt.Errorf("os.IsNotExist : %v", err)
			}
		}
		bar.Add(1)
	}
	return nil
}

func (L *Req) RequestBookData(pageNo int, pageSize int) (*BookItemsResponse, error) {
	url, err := url.Parse("http://data4library.kr/api/itemSrch")
	if err != nil {
		return nil, err
	}
	queryParam := url.Query()
	queryParam.Set("authKey", L.libAPIKey)
	queryParam.Set("libCode", strconv.Itoa(L.libCode))
	queryParam.Set("startDt", L.startDate)
	queryParam.Set("endDt", L.endDate)
	queryParam.Set("format", "json")
	queryParam.Set("pageNo", strconv.Itoa(pageNo))
	queryParam.Set("pageSize", strconv.Itoa(pageSize))

	url.RawQuery = queryParam.Encode()
	resp, err := http.Get(url.String())
	if resp.StatusCode != 200 || err != nil {
		return nil, fmt.Errorf("response Error %v status code : %v ", err, resp.StatusCode)
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
func (L *Req) Preprocess(resp *BookItemsResponse) (*[]BookItemsDoc, error) {
	var docs []BookItemsDoc
	for _, doc := range resp.Docs {
		docs = append(docs, doc.Doc)
	}
	if docs == nil {
		return nil, fmt.Errorf("docs is empty")
	}
	return &docs, nil
}
func (L *Req) Save(jsonPath string, bookItems *[]BookItemsDoc) error {
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

type DB struct {
	query    *sqlc.Queries
	libCode  int
	dataPath string
}

func NewDB(query *sqlc.Queries, libCode int, dataPath string) *DB {
	return &DB{
		query,
		libCode,
		dataPath,
	}
}

func (D *DB) InsertAll(dir string) {
	// dir := "111007/2021-01-01-2023-12-31-1000-42"

	folderPath := filepath.Join(D.dataPath, strconv.Itoa(D.libCode), dir)
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		log.Fatalln("Entries : ", err.Error())
	}

	var scrapedNum []int
	for _, entry := range entries {
		scrapNum, err := strconv.Atoi(strings.TrimSuffix(entry.Name(), ".json"))
		if err != nil {
			log.Fatalln("Atoi converting error : ", err.Error())
		}
		scrapedNum = append(scrapedNum, scrapNum)
		jsonPath := filepath.Join(folderPath, entry.Name())

		err = D.InsertBooks(jsonPath)
		if err != nil {
			log.Fatalln("InsertLibBook : ", err.Error())
		}

		err = D.InsertLibsBooks(jsonPath)
		if err != nil {
			log.Fatalln("InsertLibsBooks : ", err.Error())
		}
	}

	//check unscraped files
	splitDir := strings.Split(dir, "-")
	totalNumStr := splitDir[len(splitDir)-1]
	fmt.Printf("%v result/expect : %v/%v \n", D.libCode, len(scrapedNum), totalNumStr)
	if strconv.Itoa(len(scrapedNum)) != totalNumStr {
		totalNum, err := strconv.Atoi(totalNumStr)
		if err != nil {
			log.Fatalln("AtoI error  : ", err.Error())
		}
		for idx := range totalNum {
			if !slices.Contains(scrapedNum, idx+1) {
				fmt.Printf("not scraped : %v.json \n", idx+1)
			}
		}
	}
}

func (D *DB) InsertBooks(jsonPath string) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	var bookJson []BookItemsDoc
	err = json.Unmarshal(b, &bookJson)
	if err != nil {
		return err
	}

	var bookDB []sqlc.InsertBooksParams
	for _, book := range bookJson {
		authorRunes := []rune(book.Authors)
		book := sqlc.InsertBooksParams{
			Title:           pgtype.Text{String: book.Bookname, Valid: true},
			Author:          pgtype.Text{String: string(authorRunes[0:slices.Min([]int{512, len(authorRunes)})]), Valid: true},
			Publisher:       pgtype.Text{String: book.Publisher, Valid: true},
			Publicationyear: pgtype.Text{String: book.PublicationYear, Valid: true},
			Isbn:            pgtype.Text{String: book.ISBN13, Valid: true},
			Setisbn:         pgtype.Text{String: book.SetISBN13, Valid: true},
			Volume:          pgtype.Text{String: book.Vol, Valid: true},
			Imageurl:        pgtype.Text{Valid: true},
			Description:     pgtype.Text{Valid: true},
		}
		bookDB = append(bookDB, book)
	}

	ctx := context.Background()
	for _, book := range bookDB {
		_, err := D.query.InsertBooks(ctx, book)
		if err != nil {
			b, _ := json.Marshal(book)
			fmt.Println(string(b))
			return err
		}
	}

	return nil
}
func (D *DB) InsertLibsBooks(jsonPath string) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	var bookJson []BookItemsDoc
	err = json.Unmarshal(b, &bookJson)
	if err != nil {
		return err
	}

	var bookDB []sqlc.InsertLibsBooksParams
	for _, book := range bookJson {
		var Shelfcode string
		var Shelfname string
		var BookCode string
		arr := book.CallNumbers
		if len(arr) > 0 && arr[0].CallNumber.ShelfLocCode != "" {
			Shelfcode = book.CallNumbers[0].CallNumber.ShelfLocCode
			Shelfname = book.CallNumbers[0].CallNumber.ShelfLocName
			BookCode = book.CallNumbers[0].CallNumber.BookCode

		}
		book := sqlc.InsertLibsBooksParams{
			Libcode:   pgtype.Int4{Int32: int32(D.libCode), Valid: true},
			Isbn:      pgtype.Text{String: book.ISBN13, Valid: true},
			Classnum:  pgtype.Text{String: book.ClassNo, Valid: true},
			Bookcode:  pgtype.Text{String: BookCode, Valid: true},
			Shelfcode: pgtype.Text{String: Shelfcode, Valid: true},
			Shelfname: pgtype.Text{String: Shelfname, Valid: true},
		}
		bookDB = append(bookDB, book)
	}
	ctx := context.Background()
	for _, book := range bookDB {
		_, err := D.query.InsertLibsBooks(ctx, book)
		if err != nil {
			return err
		}
	}

	return nil
}

func (D *DB) InsertLibInfo(jsonPath string) error {
	var rawData []map[string]string
	f, err := os.Open(jsonPath)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &rawData)
	if err != nil {
		return err
	}

	var libraries []sqlc.InsertLibrariesParams
	for _, d := range rawData {

		libCodeInt, err := strconv.ParseInt(d["libCode"], 10, 32)
		if err != nil {
			libCodeInt = 0
		}

		LatitudeFloat, err := strconv.ParseFloat(d["latitude"], 64)
		if err != nil {
			return err
		}
		LongitudeFloat, err := strconv.ParseFloat(d["longitude"], 64)
		if err != nil {
			return err
		}
		var bookCountInt int
		if d["BookCount"] != "-" {
			bookCountInt, err = strconv.Atoi(d["BookCount"])
			if err != nil {
				return err
			}
		}

		library := sqlc.InsertLibrariesParams{
			Libcode:       pgtype.Int4{Int32: int32(libCodeInt), Valid: true},
			Libname:       pgtype.Text{String: d["libName"], Valid: true},
			Libaddress:    pgtype.Text{String: d["address"], Valid: true},
			Tel:           pgtype.Text{String: d["tel"], Valid: true},
			Latitude:      pgtype.Float8{Float64: LatitudeFloat, Valid: true},
			Longtitude:    pgtype.Float8{Float64: LongitudeFloat, Valid: true},
			Homepage:      pgtype.Text{String: d["homepage"], Valid: true},
			Closed:        pgtype.Text{String: d["closed"], Valid: true},
			Operatingtime: pgtype.Text{String: d["operatingTime"], Valid: true},
			Bookcount:     pgtype.Int4{Int32: int32(bookCountInt), Valid: true},
		}
		libraries = append(libraries, library)
	}
	ctx := context.Background()
	i, err := D.query.InsertLibraries(ctx, libraries)
	if err != nil {
		return err
	}
	fmt.Println(i)

	return nil
}
