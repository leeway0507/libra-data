package scrap

import (
	"libraData/config"
	"path/filepath"
	"testing"
)

func TestNaver(t *testing.T) {
	cfg := config.GetEnvConfig()
	testDataPath := filepath.Join(cfg.DATA_PATH, "test", "library_api", "naver")
	t.Run("request naver multi", func(t *testing.T) {
		isbns := []string{
			"9791163036227",
			"9791165219468",
			"9791192932477",
			"9791163031970",
		}
		RequestNaverAll(isbns, testDataPath, 2)
	})
	t.Run("request naver", func(t *testing.T) {
		isbn := "중국에 통일 제국이 서다"
		RequestNaver(isbn, testDataPath)
	})
}
