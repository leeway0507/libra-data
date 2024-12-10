package libraData

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreprocess(t *testing.T) {
	currPath, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("change libinfo dataType", func(t *testing.T) {
		// err := ChangeLibInfoDataType(filepath.Join(currPath, "data/test/libinfo-test.json"))
		err := ChangeLibInfoDataType(filepath.Join(currPath, "data/libinfo/libinfo.json"))
		if err != nil {
			t.Fatal(err)
		}
	})
}
