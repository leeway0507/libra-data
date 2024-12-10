package libraData

import (
	"libraData/model"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestUtils(t *testing.T) {
	currPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	t.Run("Test Load File", func(t *testing.T) {
		_, err = LoadFile(filepath.Join(currPath, "test_file/load_json.json"))
		if err != nil {
			log.Fatal(err)
		}
	})

	t.Run("convert csv EucKr to Utf-8", func(t *testing.T) {
		err := ConvertCsvEucKrToUtf("./data/test/euc-kr-test.csv")
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("change csv header", func(t *testing.T) {
		header := []string{"number",
			"title",
			"author",
			"publisher",
			"publication_year",
			"isbn",
			"set_isbn",
			"additional_code",
			"volume",
			"subject_code",
			"book_count",
			"loan_count",
			"registration_date",
			""}
		err := ChangeCsvHeader("./data/yangcheon-24-11-utf.csv", header)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("convert CSV to JSON", func(t *testing.T) {
		err := ConvertCsvToJson[model.Book]("./data/test/for-json-converting.csv")
		if err != nil {
			t.Fatal(err)
		}
	})
}
