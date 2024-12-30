package book_scrap

import (
	"context"
	"encoding/json"
	"io"
	"libraData/db/sqlc"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
)

var ALREADY_UPDATED = "U"

type BookScrap struct {
	query    *sqlc.Queries
	dataPath string
}

func New(conn *pgx.Conn, dataPath string) *BookScrap {
	return &BookScrap{
		query:    sqlc.New(conn),
		dataPath: dataPath,
	}
}
func (book *BookScrap) DistributeDataByIsbn(scrapPath string) error {
	entries, err := os.ReadDir(scrapPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		file, err := os.Open(filepath.Join(scrapPath, entry.Name()))
		if err != nil {
			return err
		}
		var JsonArr []sqlc.UpdateScrapDataParams
		b, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(b, &JsonArr); err != nil {
			return err
		}
		for _, jsonData := range JsonArr {
			filePath := filepath.Join(book.dataPath, jsonData.Isbn.String+".json")
			if _, err := os.Stat(filePath); err != nil {
				if os.IsNotExist(err) {
					file, err := os.Create(filePath)
					if err != nil {
						return err
					}
					defer file.Close()
					b, err := json.Marshal(jsonData)
					if err != nil {
						return err
					}
					_, err = file.Write(b)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return err
}

func (book *BookScrap) InsertToDB() error {
	entries, err := os.ReadDir(book.dataPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fileName := entry.Name()
		_, isJson := strings.CutSuffix(fileName, ".json")
		if !isJson {
			log.Println("found not json file", fileName)
			continue
		}
		if fileName[:1] == ALREADY_UPDATED {
			log.Printf("%v already updated \n", fileName)
			continue
		}
		file, err := os.Open(filepath.Join(book.dataPath, fileName))
		if err != nil {
			return err
		}
		var JsonArr sqlc.UpdateScrapDataParams
		b, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(b, &JsonArr); err != nil {
			return err
		}
		ctx := context.Background()

		if err = book.query.UpdateScrapData(ctx, JsonArr); err != nil {
			return err
		}

		err = os.Rename(filepath.Join(book.dataPath, fileName),
			filepath.Join(book.dataPath, "U"+fileName))
		if err != nil {
			return err
		}

	}
	return nil
}
