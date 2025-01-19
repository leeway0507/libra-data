package handler

import (
	"context"
	"encoding/json"
	"io"
	"libraData/pkg/db"
	"libraData/pkg/db/sqlc"
	"libraData/pkg/pb"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/proto"
)

var (
	collectYear = []string{
		"2014",
		"2015",
		"2016",
		"2017",
		"2018",
		"2019",
		"2020",
		"2021",
		"2022",
		"2023",
		"2024"}

	defaultColName = []string{
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
)

func Insert(db_url string, dataPath string, scrapDate string, workers int) {
	ctx := context.Background()
	pool := db.ConnectPGPool(db_url, ctx)
	defer pool.Close()

	folders := LoadLibNaruFolders(dataPath)

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
	dbInstance := NewLibBooksHandler(query, scrapData, dataPath)
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

func LoadLibNaruFolders(dataPath string) []os.DirEntry {
	entries, err := os.ReadDir(dataPath)
	if err != nil {
		panic(err)
	}
	if len(entries) == 0 {
		panic("no lib folders")
	}
	return entries
}

// scrap data handler
type sh struct {
	query     *sqlc.Queries
	dataPath  string
	scrapDate string
}

func NewLibBooksHandler(query *sqlc.Queries, scrapDate string, dataPath string) *sh {
	return &sh{
		query:     query,
		dataPath:  dataPath,
		scrapDate: scrapDate,
	}
}

func (S *sh) Load() (*pb.BookRows, error) {
	file, err := os.Open(filepath.Join(S.dataPath, S.scrapDate+".pb"))
	if err != nil {
		log.Println("Load :", err)
		return nil, nil
	}

	defer file.Close()
	bookRows := &pb.BookRows{}
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

func (S *sh) InsertBooks(books []*pb.BookRow) error {
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
		_, err := S.query.InsertBooks(ctx, book)
		if err != nil {
			b, _ := json.Marshal(book)
			log.Println(string(b))
			return err
		}
	}

	return nil
}
func (S *sh) InsertLibsBooks(books []*pb.BookRow, libCode string) error {
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
		_, err := S.query.InsertLibsBooks(ctx, book)
		if err != nil {
			return err
		}
	}

	return nil
}

func (S *sh) MarkAsUpdated() error {
	err := os.Rename(filepath.Join(S.dataPath, S.scrapDate+".pb"),
		filepath.Join(S.dataPath, "U"+S.scrapDate+".pb"))
	if err != nil {
		return err
	}
	return nil
}
