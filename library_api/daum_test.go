package library_api

import (
	"libraData/config"
	"path/filepath"
	"testing"
)

func TestDaum(t *testing.T) {
	cfg := config.GetEnvConfig()
	testDataPath := filepath.Join(cfg.DATA_PATH, "test", "library_api", "daum")
	t.Run("request daum multi", func(t *testing.T) {
		isbns := []string{
			"9791163036227",
			"9791165219468",
			"9791192932477",
			"9791163031970",
		}
		RequestDaumMultiple(isbns, testDataPath, 2)
	})
	t.Run("single daum request", func(t *testing.T) {
		isbn := "9791163034735"
		RequestDaum(isbn, testDataPath)
	})
}
