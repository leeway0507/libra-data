package libraData

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

func LoadFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return raw, err

}

func MakeFolders(path string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}
	return nil
}

func ConvertCsvEucKrToUtf(path string) error {
	readFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer readFile.Close()

	koreanTransformReader := transform.NewReader(readFile, korean.EUCKR.NewDecoder())
	csvReader := csv.NewReader(koreanTransformReader)

	if err := MakeFolders(path); err != nil {
		return err
	}

	writeFileName := strings.TrimSuffix(path, ".csv") + "-utf.csv"
	writeFile, err := os.Create(writeFileName)
	if err != nil {
		return err
	}
	defer writeFile.Close()

	csvWriter := csv.NewWriter(writeFile)
	defer csvWriter.Flush()

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func ChangeCsvHeader(path string, newHeader []string) error {
	file, err := os.Open(path)
	if nil != err {
		return err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	oldHeader, err := csvReader.Read()

	if nil != err {
		return err
	}
	if len(oldHeader) != len(newHeader) {
		return fmt.Errorf("length doens't mathch \n old header length: %v \n new header length %v",
			len(oldHeader), len(newHeader))
	}
	tempFile, err := os.Create(filepath.Join(filepath.Dir(path), "temp.csv"))
	if err != nil {
		return err
	}

	csvWriter := csv.NewWriter(tempFile)
	csvWriter.Write(newHeader)
	defer csvWriter.Flush()

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := csvWriter.Write(record); err != nil {
			return err
		}
	}

	if err := os.Rename(tempFile.Name(), path); err != nil {
		return err
	}
	return nil
}

func ConvertCsvToJson[T any](path string) error {
	file, err := os.Open(path)
	if nil != err {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()

	if nil != err {
		return err
	}
	var jsonArray []T
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if nil != err {
			return err
		}

		item := new(T)
		fields := reflect.ValueOf(item).Elem()
		fieldKeys := reflect.TypeOf(item).Elem()

		for idx := range len(header) - 1 {
			field := fields.Field(idx)
			fieldKey := fieldKeys.Field(idx).Name
			valueType := field.Kind()

			switch valueType {
			case reflect.Int:
				var intValue int64
				if record[idx] != "" {
					intValue, err = strconv.ParseInt(record[idx], 10, 64)
					if err != nil {
						return fmt.Errorf(err.Error(), "key : ", fieldKey, "idx : ", record[0])
					}
				}
				field.SetInt(intValue)

			case reflect.String:
				field.SetString(record[idx])

			case reflect.Struct:
				if field.Type() == reflect.TypeOf(time.Time{}) {
					parsedTime, err := time.Parse("2006-01-02", record[idx])
					if err != nil {
						return fmt.Errorf(err.Error(), "key : ", fieldKey, "idx : ", record[0])
					}
					field.Set(reflect.ValueOf(parsedTime))
				}
			}
		}
		jsonArray = append(jsonArray, *item)
	}

	jsonFile, err := os.Create(strings.TrimSuffix(path, ".csv") + ".json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	encoder := json.NewEncoder(jsonFile)
	encoder.Encode(jsonArray)
	return nil
}
