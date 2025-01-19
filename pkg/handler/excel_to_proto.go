package handler

import (
	"fmt"
	"libraData/pkg/pb"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sync"

	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/proto"
)

// 정보나루에서 수집한 xlsx 파일을 proto buffer로 변환
func ConvertExcelToProto(scrapDate string, dataPath string, workers int) error {
	folders := LoadLibNaruFolders(dataPath)

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
				requiresProcessing := ep.IsConvertingRequired()
				log.Println(folder.Name(), "requiresProcessing", requiresProcessing)
				if requiresProcessing {
					err := ep.Convert()
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

func (ep *excelToProto) IsConvertingRequired() bool {
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

func (ep *excelToProto) Convert() error {
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
