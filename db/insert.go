package db

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	sqlc "libraData/db/sqlc"
	"libraData/utils"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func InsertLibBookBulkFromCSV(conn *sqlc.Queries, ctx context.Context, csvPath string) error {
	headerOrder := []string{"번호", "도서명", "저자", "출판사", "발행년도", "ISBN", "세트 ISBN", "부가기호", "권", "주제분류번호", "도서권수", "대출건수", "등록일자", ""}
	f, err := os.Open(csvPath)

	if err != nil {
		return err
	}
	csvReader := csv.NewReader(f)
	currHeaderOrder, _ := csvReader.Read() // remove header

	for idx := range headerOrder {
		if headerOrder[idx] != currHeaderOrder[idx] {
			return fmt.Errorf("header order not matched: %s", currHeaderOrder)
		}
	}

	var books []sqlc.InsertBooksParams
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		bookCountInt, err := strconv.ParseInt(record[10], 10, 32)
		if err != nil {
			bookCountInt = 0
		}

		loanCountInt, err := strconv.ParseInt(record[11], 10, 32)
		if err != nil {
			loanCountInt = 0
		}

		// registrationDate 파싱
		registrationDate, err := time.Parse("2006-01-02", record[12])
		if err != nil {
			return fmt.Errorf("invalid date format for registration date: %v", record[12])
		}

		book := sqlc.InsertBooksParams{
			// {"번호", "도서명", "저자", "출판사", "발행년도", "ISBN", "세트 ISBN", "부가기호", "권", "주제분류번호", "도서권수", "대출건수", "등록일자", ""}
			Title:            pgtype.Text{String: record[1], Valid: true},
			Author:           pgtype.Text{String: record[2], Valid: true},
			Publisher:        pgtype.Text{String: record[3], Valid: true},
			Publicationyear:  pgtype.Text{String: record[4], Valid: true},
			Isbn:             pgtype.Text{String: record[5], Valid: true},
			Setisbn:          pgtype.Text{String: record[6], Valid: true},
			Additionalcode:   pgtype.Text{String: record[7], Valid: true},
			Volume:           pgtype.Text{String: record[8], Valid: true},
			Subjectcode:      pgtype.Text{String: record[9], Valid: true},
			Bookcount:        pgtype.Int4{Int32: int32(bookCountInt), Valid: true},
			Loancount:        pgtype.Int4{Int32: int32(loanCountInt), Valid: true},
			Registrationdate: pgtype.Date{Time: registrationDate, Valid: true},
		}
		books = append(books, book)
	}

	i, err := conn.InsertBooks(ctx, books)
	if err != nil {
		return err
	}
	fmt.Println(i)

	return nil
}

func InsertLibInfoBulkFromJSON(conn *sqlc.Queries, ctx context.Context, jsonPath string) error {
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

	i, err := conn.InsertLibraries(ctx, libraries)
	if err != nil {
		return err
	}
	fmt.Println(i)

	return nil
}
