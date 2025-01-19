package handler

import (
	"context"
	"encoding/json"
	"io"
	"libraData/pkg/db/sqlc"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

var (
	UPDATED        = "U"
	NOT_EXIST_ISBN = "N"
)

type bookSpec struct {
	query    *sqlc.Queries
	dataPath string
}

func NewBookSpec(query *sqlc.Queries, dataPath string) *bookSpec {
	return &bookSpec{
		query,
		dataPath,
	}
}

func (bs *bookSpec) SeparateScrapDataByIsbn(scrapPath string) error {
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

			filePath := filepath.Join(bs.dataPath, jsonData.Isbn.String+".json")
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

func (bs *bookSpec) InsertScrapData() error {
	entries, err := os.ReadDir(bs.dataPath)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		fileName := entry.Name()
		_, isJson := strings.CutSuffix(fileName, ".json")
		if !isJson {
			log.Println("found non json file", fileName)
			continue
		}
		if fileName[:1] == UPDATED {
			// log.Printf("%v already updated \n", fileName)
			continue
		}
		file, err := os.Open(filepath.Join(bs.dataPath, fileName))
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
		if err = bs.query.UpdateScrapData(ctx, JsonArr); err != nil {
			return err
		}

		err = os.Rename(filepath.Join(bs.dataPath, fileName),
			filepath.Join(bs.dataPath, "U"+fileName))
		if err != nil {
			return err
		}

	}
	return nil
}

func (bs *bookSpec) InsertToc() error {
	isbnPath := filepath.Join(bs.dataPath, "toc")
	entries, err := os.ReadDir(isbnPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fileName := entry.Name()
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
		if err = bs.query.UpdateToc(ctx, jsonArr); err != nil {
			return err
		}

		err = os.Rename(filepath.Join(isbnPath, fileName),
			filepath.Join(isbnPath, UPDATED+fileName))
		if err != nil {
			return err
		}
	}
	return nil
}

// 도서 정보는 존재하나 목차는 없는 도서의 경우 Toc를 빈값으로 저장
func (bs *bookSpec) InsertEmptyToc() error {
	isbnPath := filepath.Join(bs.dataPath, "toc")
	entries, err := os.ReadDir(isbnPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fileName := entry.Name()
		ctx := context.Background()
		if !slices.Contains([]string{NOT_EXIST_ISBN}, fileName[:1]) {
			return nil
		}
		fileName, _ = strings.CutSuffix(fileName, ".json")
		err := bs.query.UpdateToc(ctx, sqlc.UpdateTocParams{
			Toc:  pgtype.Text{String: "", Valid: true},
			Isbn: pgtype.Text{String: fileName[1:], Valid: true},
		})
		if err != nil {
			log.Printf("err: %#+v\n", err)
		}

	}
	return nil
}
