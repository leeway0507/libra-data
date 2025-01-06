package utils

import (
	"log"
	"os"
	"path/filepath"
)

func ResetUpdateStatus(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		fileName := entry.Name()

		if fileName[:1] != "U" {
			continue
		}
		err = os.Rename(filepath.Join(path, fileName),
			filepath.Join(path, fileName[1:]))
		if err != nil {
			panic(err)
		}
	}
}

func HandleErr(err error, msg string) {
	if err != nil {
		log.Printf("%s Error: %v", msg, err)
		return
	}
}

func CheckFileExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
