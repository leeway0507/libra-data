package db

import (
	"context"
	"encoding/json"
	"io"
	sqlc "libraData/db/sqlc"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var ALREADY_UPDATED = "U"

func UpdateScrapDataFromJson(query *sqlc.Queries, ctx context.Context) error {
	isbnPath := filepath.Join(cfg.DATA_PATH, "isbn")
	entries, err := os.ReadDir(isbnPath)
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
		file, err := os.Open(filepath.Join(isbnPath, fileName))
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
		if err = query.UpdateScrapData(ctx, JsonArr); err != nil {
			return err
		}
		err = os.Rename(filepath.Join(isbnPath, fileName),
			filepath.Join(isbnPath, "U"+fileName))
		if err != nil {
			return err
		}

	}
	return nil
}

func DistributeScrapDatasByIsbn(scrapPath string, targetPath string) error {
	// load scrap file
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
			filePath := filepath.Join(targetPath, jsonData.Isbn.String+".json")
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

func UpdateEmbeddingFromPB(query *sqlc.Queries, ctx context.Context) error {
	isbnPath := filepath.Join(cfg.DATA_PATH, "embedding")
	entries, err := os.ReadDir(isbnPath)
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
		file, err := os.Open(filepath.Join(isbnPath, fileName))
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
		if err = query.UpdateScrapData(ctx, JsonArr); err != nil {
			return err
		}
		err = os.Rename(filepath.Join(isbnPath, fileName),
			filepath.Join(isbnPath, "U"+fileName))
		if err != nil {
			return err
		}
	}
	return nil
}
