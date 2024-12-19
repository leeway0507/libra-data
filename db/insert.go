package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"libraData/collect"
	sqlc "libraData/db/sqlc"
	"libraData/pb"
	"libraData/utils"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/exp/slices"
)

func InsertLibBookBulkFromJSON(query *sqlc.Queries, ctx context.Context, jsonPath string) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	var bookJson []collect.BookItemsDoc
	err = json.Unmarshal(b, &bookJson)
	if err != nil {
		return err
	}

	var bookDB []sqlc.InsertBooksParams
	for _, book := range bookJson {
		authorRunes := []rune(book.Authors)
		book := sqlc.InsertBooksParams{
			Title:           pgtype.Text{String: book.Bookname, Valid: true},
			Author:          pgtype.Text{String: string(authorRunes[0:slices.Min([]int{512, len(authorRunes)})]), Valid: true},
			Publisher:       pgtype.Text{String: book.Publisher, Valid: true},
			Publicationyear: pgtype.Text{String: book.PublicationYear, Valid: true},
			Isbn:            pgtype.Text{String: book.ISBN13, Valid: true},
			Setisbn:         pgtype.Text{String: book.SetISBN13, Valid: true},
			Volume:          pgtype.Text{String: book.Vol, Valid: true},
			Imageurl:        pgtype.Text{Valid: true},
			Description:     pgtype.Text{Valid: true},
		}
		bookDB = append(bookDB, book)
	}

	for _, book := range bookDB {
		_, err := query.InsertBooks(ctx, book)
		if err != nil {
			b, _ := json.Marshal(book)
			fmt.Println(string(b))
			return err
		}
	}

	return nil
}
func InsertLibBookBulkFromPB(query *sqlc.Queries, ctx context.Context, books []*pb.BookRow) error {
	var bookDB []sqlc.InsertBooksParams
	for _, book := range books {
		authorRunes := []rune(book.Author)
		book := sqlc.InsertBooksParams{
			Title:           pgtype.Text{String: book.Title, Valid: true},
			Author:          pgtype.Text{String: string(authorRunes[0:slices.Min([]int{512, len(authorRunes)})]), Valid: true},
			Publisher:       pgtype.Text{String: book.Publisher, Valid: true},
			Publicationyear: pgtype.Text{String: book.PublicationYear, Valid: true},
			Isbn:            pgtype.Text{String: book.Isbn, Valid: true},
			Setisbn:         pgtype.Text{String: book.SetIsbn, Valid: true},
			Volume:          pgtype.Text{String: book.Volume, Valid: true},
			Imageurl:        pgtype.Text{Valid: true},
			Description:     pgtype.Text{Valid: true},
		}
		bookDB = append(bookDB, book)
	}

	for _, book := range bookDB {
		_, err := query.InsertBooks(ctx, book)
		if err != nil {
			b, _ := json.Marshal(book)
			fmt.Println(string(b))
			return err
		}
	}

	return nil
}

func InsertLibsBooksRelationBulkFromJSON(query *sqlc.Queries, ctx context.Context, jsonPath string, libCode int32) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	var bookJson []collect.BookItemsDoc
	err = json.Unmarshal(b, &bookJson)
	if err != nil {
		return err
	}

	var bookDB []sqlc.InsertLibsBooksParams
	for _, book := range bookJson {
		var Shelfcode string
		var Shelfname string
		var BookCode string
		arr := book.CallNumbers
		if len(arr) > 0 && arr[0].CallNumber.ShelfLocCode != "" {
			Shelfcode = book.CallNumbers[0].CallNumber.ShelfLocCode
			Shelfname = book.CallNumbers[0].CallNumber.ShelfLocName
			BookCode = book.CallNumbers[0].CallNumber.BookCode

		}
		book := sqlc.InsertLibsBooksParams{
			Libcode:   pgtype.Int4{Int32: libCode, Valid: true},
			Isbn:      pgtype.Text{String: book.ISBN13, Valid: true},
			Classnum:  pgtype.Text{String: book.ClassNo, Valid: true},
			Bookcode:  pgtype.Text{String: BookCode, Valid: true},
			Shelfcode: pgtype.Text{String: Shelfcode, Valid: true},
			Shelfname: pgtype.Text{String: Shelfname, Valid: true},
		}
		bookDB = append(bookDB, book)
	}

	for _, book := range bookDB {
		_, err := query.InsertLibsBooks(ctx, book)
		if err != nil {
			return err
		}
	}

	return nil
}
func InsertLibsBooksRelationBulkFromPB(query *sqlc.Queries, ctx context.Context, books []*pb.BookRow, libCode int32) error {

	var bookDB []sqlc.InsertLibsBooksParams
	for _, book := range books {
		book := sqlc.InsertLibsBooksParams{
			Libcode:  pgtype.Int4{Int32: libCode, Valid: true},
			Isbn:     pgtype.Text{String: book.Isbn, Valid: true},
			Classnum: pgtype.Text{String: book.ClassNum, Valid: true},
		}
		bookDB = append(bookDB, book)
	}

	for _, book := range bookDB {
		_, err := query.InsertLibsBooks(ctx, book)
		if err != nil {
			return err
		}
	}

	return nil
}
func InsertLibInfoBulkFromJSON(query *sqlc.Queries, ctx context.Context, jsonPath string) error {
	var rawData []map[string]string
	file, err := utils.LoadFile(jsonPath)

	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &rawData)
	if err != nil {
		return err
	}

	var libraries []sqlc.InsertLibrariesParams
	for _, d := range rawData {

		libCodeInt, err := strconv.ParseInt(d["libCode"], 10, 32)
		if err != nil {
			libCodeInt = 0
		}

		LatitudeFloat, err := strconv.ParseFloat(d["latitude"], 64)
		if err != nil {
			return err
		}
		LongitudeFloat, err := strconv.ParseFloat(d["longitude"], 64)
		if err != nil {
			return err
		}
		var bookCountInt int
		if d["BookCount"] != "-" {
			bookCountInt, err = strconv.Atoi(d["BookCount"])
			if err != nil {
				return err
			}
		}

		library := sqlc.InsertLibrariesParams{
			Libcode:       pgtype.Int4{Int32: int32(libCodeInt), Valid: true},
			Libname:       pgtype.Text{String: d["libName"], Valid: true},
			Libaddress:    pgtype.Text{String: d["address"], Valid: true},
			Tel:           pgtype.Text{String: d["tel"], Valid: true},
			Latitude:      pgtype.Float8{Float64: LatitudeFloat, Valid: true},
			Longtitude:    pgtype.Float8{Float64: LongitudeFloat, Valid: true},
			Homepage:      pgtype.Text{String: d["homepage"], Valid: true},
			Closed:        pgtype.Text{String: d["closed"], Valid: true},
			Operatingtime: pgtype.Text{String: d["operatingTime"], Valid: true},
			Bookcount:     pgtype.Int4{Int32: int32(bookCountInt), Valid: true},
		}
		libraries = append(libraries, library)
	}

	i, err := query.InsertLibraries(ctx, libraries)
	if err != nil {
		return err
	}
	fmt.Println(i)

	return nil
}
