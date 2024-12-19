package preprocess

import (
	"fmt"
	"io"
	"libraData/config"
	"libraData/pb"
	"os"
	"path/filepath"
	"reflect"
	"slices"

	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/proto"
)

var cfg *config.EnvConfig = config.GetEnvConfig()
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

type ExcelToProto struct {
	entry      os.DirEntry
	scrapDate  string
	folderPath string
}

func ConvertExcelToProto(scrapDate string) error {
	libEntries := LoadLibScraperFolder()
	if len(libEntries) == 0 {
		return fmt.Errorf("no lib folders")
	}

	for _, libEntry := range libEntries {
		ep := NewExcelToProto(libEntry, scrapDate)
		isPreprocessed := ep.GetPreprocessStatus()
		if isPreprocessed {
			continue
		}
		err := ep.Preprocess()
		if err != nil {
			return err
		}
	}
	return nil
}

func NewExcelToProto(entry os.DirEntry, scrapDate string) *ExcelToProto {
	return &ExcelToProto{
		entry:      entry,
		scrapDate:  scrapDate,
		folderPath: filepath.Join(cfg.DATA_PATH, "library", entry.Name()),
	}
}

func LoadLibScraperFolder() []os.DirEntry {
	entries, err := os.ReadDir(filepath.Join(cfg.DATA_PATH, "library"))
	if err != nil {
		panic(err)
	}
	return entries

}

func (ep *ExcelToProto) GetPreprocessStatus() bool {
	if !ep.entry.IsDir() {
		fmt.Printf("Unintened file %s \n", ep.entry.Name())
		return true
	}
	files, err := os.ReadDir(ep.folderPath)
	if err != nil {
		fmt.Printf("file does not exist in %s \n", ep.entry.Name())
		return true
	}

	for _, file := range files {
		if file.Name() == ep.scrapDate+".pb" {
			return true
		}
		if file.Name() == ep.scrapDate+".xlsx" {
			return false
		}
	}
	fmt.Printf("%s does not exist in %s \n", ep.scrapDate+".xlsx", ep.entry.Name())
	return true
}

func (ep *ExcelToProto) Preprocess() error {
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
	f, err := os.Create(filepath.Join(ep.folderPath, ep.scrapDate+".pb"))
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

func (ep *ExcelToProto) LoadBooksFromPB() (*pb.BookRows, error) {
	file, err := os.Open(filepath.Join(ep.folderPath, ep.scrapDate+".pb"))
	if err != nil {
		return nil, err
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

func (ep *ExcelToProto) LoadExcelRows() ([][]string, error) {
	f, err := excelize.OpenFile(filepath.Join(ep.folderPath, ep.scrapDate+".xlsx"))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
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
