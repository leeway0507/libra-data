package library_scrap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"libraData/db"
	"libraData/db/sqlc"
	"libraData/pb"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/proto"
)

var collectYear = []string{"2014", "2015", "2016", "2017", "2018", "2019", "2020", "2021", "2022", "2023", "2024"}
var defaultColName []string = []string{
	"번호",  //0
	"도서명", //1
	"저자",  //2
	"출판사",
	"발행년도",    //4
	"ISBN",    //5
	"세트 ISBN", //6
	"부가기호",
	"권",
	"주제분류번호", //9
	"도서권수",   //10
	"대출건수",
	"등록일자",
}

func ConvertExcelToProto(scrapDate string, dataPath string, workers int) error {
	folders := LoadLibScrapFolders(dataPath)

	// 채널을 통해 작업을 전달
	tasks := make(chan os.DirEntry, len(folders))
	errChan := make(chan error, workers)
	var wg sync.WaitGroup

	// 워커 풀 생성
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for folder := range tasks {
				// log.Println(folder.Name(), "에 대해 작업 중..")
				ep := NewExcelToProto(folder, scrapDate, dataPath)
				requiresProcessing := ep.IsPreprocessingRequired()
				log.Println(folder.Name(), "requiresProcessing", requiresProcessing)
				if requiresProcessing {
					err := ep.Preprocess()
					if err != nil {
						errChan <- err
					}
				}
			}
		}()
	}

	// 작업 채널에 작업 추가
	for _, folder := range folders {
		tasks <- folder
	}
	// 종료 신호
	close(tasks)

	wg.Wait()
	return nil
}

func InsertAll(db_url string, dataPath string, scrapDate string, workers int) {

	ctx := context.Background()
	pool := db.ConnectPGPool(db_url, ctx)
	defer pool.Close()

	folders := LoadLibScrapFolders(dataPath)

	tasks := make(chan os.DirEntry, len(folders))

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			conn := db.ConnectPG(db_url, ctx)
			defer conn.Close(ctx)

			q := sqlc.New(conn)
			for folder := range tasks {
				log.Println(workerID+1, "worker", folder.Name())
				if !folder.IsDir() {
					log.Printf("%s is not a dir. \n", folder.Name())
					continue
				}
				specDataPath := filepath.Join(dataPath, folder.Name())
				insert(q, scrapDate, specDataPath)
			}
		}(i)
	}
	// 작업 채널에 작업 추가
	for _, folder := range folders {
		tasks <- folder
	}
	// 종료 신호
	close(tasks)
	wg.Wait()
}

func insert(query *sqlc.Queries, scrapData string, dataPath string) {
	ctx := context.Background()
	dbInstance := NewDB(query, scrapData, dataPath)
	data, err := dbInstance.Load()
	if err != nil {
		log.Fatalln(err)
	}
	if data == nil {
		return
	}
	err = dbInstance.InsertBooks(data.Books)
	if err != nil {
		log.Fatalln("InsertBooks : ", err.Error())
	}
	libCode, err := query.GetLibCodFromLibName(ctx, pgtype.Text{String: filepath.Base(dataPath), Valid: true})
	if err != nil {
		log.Fatalln("GetLibCodFromLibName : ", err.Error())
	}
	err = dbInstance.InsertLibsBooks(data.Books, libCode.String)
	if err != nil {
		log.Fatalln("GetLibCodFromLibName : ", err.Error())
	}
	err = dbInstance.MarkAsUpdated()
	if err != nil {
		log.Fatalln("GetLibCodFromLibName : ", err.Error())
	}

}

type excelToProto struct {
	entry     os.DirEntry
	scrapDate string
	dataPath  string
}

func NewExcelToProto(entry os.DirEntry, scrapDate string, dataPath string) *excelToProto {
	return &excelToProto{
		entry:     entry,
		scrapDate: scrapDate,
		dataPath:  filepath.Join(dataPath, entry.Name()),
	}
}

func (ep *excelToProto) IsPreprocessingRequired() bool {
	if !ep.entry.IsDir() {
		fmt.Printf("Unintened file %s \n", ep.entry.Name())
		return false
	}
	folder, err := os.ReadDir(ep.dataPath)
	if err != nil {
		fmt.Printf("file does not exist in %s \n", ep.entry.Name())
		return false
	}

	for _, file := range folder {
		pbList := []string{ep.scrapDate + ".pb", "U" + ep.scrapDate + ".pb"}
		if slices.Contains(pbList, file.Name()) {
			// err := os.Remove(filepath.Join(ep.dataPath, ep.scrapDate+".pb"))
			// if err != nil {
			// 	log.Println(err)
			// }
			// log.Println("removed")
			return false
		}
	}
	for _, file := range folder {
		if file.Name() == ep.scrapDate+".xlsx" {
			return true
		}
	}
	fmt.Printf("%s does not exist in %s \n", ep.scrapDate+".xlsx", ep.entry.Name())
	return false
}

func (ep *excelToProto) Preprocess() error {
	rows, err := ep.LoadExcelRows()
	if err != nil {
		return err
	}
	var pbRows pb.BookRows
	for idx, row := range rows {
		if idx == 0 {
			continue
		}
		pubYear := row[4]
		if slices.Contains(collectYear, pubYear) {
			pbRow := &pb.BookRow{
				Title:           row[1],
				Author:          row[2],
				Publisher:       row[3],
				PublicationYear: row[4],
				Isbn:            row[5],
				SetIsbn:         row[7],
				ClassNum:        row[9],
				Volume:          row[10],
			}
			pbRows.Books = append(pbRows.Books, pbRow)
		}
	}
	b, err := proto.Marshal(&pbRows)
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(ep.dataPath, ep.scrapDate+".pb"))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (ep *excelToProto) LoadExcelRows() ([][]string, error) {
	f, err := excelize.OpenFile(filepath.Join(ep.dataPath, ep.scrapDate+".xlsx"))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println(err)
		}
	}()
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	colName := rows[0]

	if !(reflect.DeepEqual(colName, defaultColName)) {
		return nil, fmt.Errorf("header is not equal \n Required:%v \n Get:%v", defaultColName, colName)
	}

	return rows, nil
}

type DB struct {
	query     *sqlc.Queries
	dataPath  string
	scrapDate string
}

func NewDB(query *sqlc.Queries, scrapDate string, dataPath string) *DB {
	return &DB{
		query:     query,
		dataPath:  dataPath,
		scrapDate: scrapDate,
	}
}

func (D *DB) Load() (*pb.BookRows, error) {
	file, err := os.Open(filepath.Join(D.dataPath, D.scrapDate+".pb"))
	if err != nil {
		log.Println("Load :", err)
		return nil, nil
	}

	defer file.Close()
	bookRows := &pb.BookRows{} // 포인터 초기화
	b, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(b, bookRows)
	if err != nil {
		return nil, err
	}
	return bookRows, nil

}

func (D *DB) InsertBooks(books []*pb.BookRow) error {
	ctx := context.Background()
	var bookDB []sqlc.InsertBooksParams
	for _, book := range books {
		authorRunes := []rune(book.Author)
		book := sqlc.InsertBooksParams{
			Title:           pgtype.Text{String: book.Title, Valid: true},
			Author:          pgtype.Text{String: string(authorRunes[0:slices.Min([]int{512, len(authorRunes)})]), Valid: true},
			Publisher:       pgtype.Text{String: book.Publisher, Valid: true},
			PublicationYear: pgtype.Text{String: book.PublicationYear, Valid: true},
			Isbn:            pgtype.Text{String: book.Isbn, Valid: true},
			Volume:          pgtype.Text{String: book.Volume, Valid: true},
			ImageUrl:        pgtype.Text{Valid: true},
			Description:     pgtype.Text{Valid: true},
		}
		bookDB = append(bookDB, book)
	}

	for _, book := range bookDB {
		_, err := D.query.InsertBooks(ctx, book)
		if err != nil {
			b, _ := json.Marshal(book)
			log.Println(string(b))
			return err
		}
	}

	return nil
}
func (D *DB) InsertLibsBooks(books []*pb.BookRow, libCode string) error {
	ctx := context.Background()
	var bookDB []sqlc.InsertLibsBooksParams
	for _, book := range books {
		book := sqlc.InsertLibsBooksParams{
			LibCode:  pgtype.Text{String: libCode, Valid: true},
			Isbn:     pgtype.Text{String: book.Isbn, Valid: true},
			ClassNum: pgtype.Text{String: book.ClassNum, Valid: true},
		}
		bookDB = append(bookDB, book)
	}

	for _, book := range bookDB {
		_, err := D.query.InsertLibsBooks(ctx, book)
		if err != nil {
			return err
		}
	}

	return nil
}

func (D *DB) MarkAsUpdated() error {
	err := os.Rename(filepath.Join(D.dataPath, D.scrapDate+".pb"),
		filepath.Join(D.dataPath, "U"+D.scrapDate+".pb"))
	if err != nil {
		return err
	}
	return nil
}

func LoadLibScrapFolders(dataPath string) []os.DirEntry {
	entries, err := os.ReadDir(dataPath)
	if err != nil {
		panic(err)
	}
	if len(entries) == 0 {
		panic("no lib folders")
	}
	return entries
}
