package collect

import (
	"testing"
)

func TestCollectData(t *testing.T) {

	var bookItemsResp *BookItemsResponse
	var bookItems *[]BookItemsDoc
	// t.Run("Get Lib Books All", func(t *testing.T) {
	// 	err := GetBookItemsAll(111015, "2024-01-01", "2024-11-30", "temp.json")
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// })
	t.Run("Get and preprocess and save lib", func(t *testing.T) {
		resp, err := GetBookItems(111015, "2024-01-01", "2024-11-30", 1, 500)
		if err != nil {
			t.Fatal(err)
		}
		docs, err := PreprocessBookItems(resp)
		if err != nil {
			t.Fatal(err)
		}
		err = SaveBookItemsAsJson("temp.json", docs)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("Get Lib books", func(t *testing.T) {
		// today := time.Now().Format(time.DateOnly)
		_, err := GetBookItems(111015, "2024-01-01", "2024-11-30", 1, 500)
		if err != nil {
			t.Fatal(err)
		}

		// fmt.Println("bookItemsResp", bookItemsResp)
	})

	t.Run("Preprocess lib books", func(t *testing.T) {
		_, err := PreprocessBookItems(bookItemsResp)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("save lib books", func(t *testing.T) {
		err := SaveBookItemsAsJson("temp.json", bookItems)
		if err != nil {
			t.Fatal(err)
		}
	})

}
