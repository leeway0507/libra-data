package main

import (
	"context"
	"libraData/config"
	"libraData/db"
	"libraData/db/sqlc"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	cfg := config.GetEnvConfig()
	ctx := context.Background()
	conn := db.ConnectPG(cfg.DATABASE_URL, ctx)
	query := sqlc.New(conn)

	folderDir := []string{
		"111003/2023-01-01-2023-12-31-500-18",
		"111003/2024-01-01-2024-12-01-500-17",
		"111004/2023-01-01-2023-12-31-500-28",
		"111004/2024-01-01-2024-12-01-500-21",
		"111006/2023-01-01-2023-12-31-500-17",
		"111006/2024-01-01-2024-12-01-500-9",
	}
	for _, dir := range folderDir {
		folderPath := filepath.Join(cfg.DATA_PATH, "temp", dir)
		entries, err := os.ReadDir(folderPath)
		if err != nil {
			log.Fatalln("Entries : ", err.Error())
		}
		for _, entry := range entries {
			jsonPath := filepath.Join(folderPath, entry.Name())
			err := db.InsertLibBookBulkFromJSON(query, ctx, jsonPath)
			if err != nil {
				log.Fatalln("InsertLibBookBulkFromJSON : ", err.Error())
			}
			libCode, err := strconv.ParseInt(strings.Split(dir, "/")[0], 10, 32)
			if err != nil {
				log.Fatalln("ParseInt : ", err.Error())
			}
			err = db.InsertLibsBooksRelationBulkFromJSON(query, ctx, jsonPath, int32(libCode))
			if err != nil {
				log.Fatalln("InsertLibsBooksRelationBulkFromJSON : ", err.Error())
			}

		}
	}
}

// func main() {
// 	cfg := config.GetEnvConfig()
// 	ctx := context.Background()
// 	conn := db.ConnectPG(cfg.DATABASE_URL, ctx)
// 	query := sqlc.New(conn)
// 	err := db.InsertLibInfoBulkFromJSON(query, ctx, filepath.Join(cfg.DATA_PATH, "libinfo", "libinfo.json"))
// 	if err != nil {
// 		log.Fatalln("InsertLibInfoBulkFromJSON : ", err.Error())
// 	}
// }
