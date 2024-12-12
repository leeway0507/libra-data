package preprocess

import (
	"encoding/json"
	"libraData/model"
	"libraData/utils"
	"os"
	"strconv"
	"strings"
)

func ChangeLibInfoDataType(path string) error {
	var rawData []map[string]string
	file, err := utils.LoadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &rawData)
	if err != nil {
		return err
	}
	var changedData []model.Lib

	for _, d := range rawData {
		var libCodeInt int
		if d["libCode"] != "" {

			libCodeInt, err = strconv.Atoi(d["libCode"])
			if err != nil {
				return err
			}
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

		libData := model.Lib{
			LibCode:       libCodeInt,
			LibName:       d["libName"],
			Address:       d["address"],
			Tel:           d["tel"],
			Fax:           d["fax"],
			Latitude:      LatitudeFloat,
			Longitude:     LongitudeFloat,
			Homepage:      d["homepage"],
			Closed:        d["closed"],
			OperatingTime: d["operatingTime"],
			BookCount:     bookCountInt,
		}

		changedData = append(changedData, libData)

		newJson, err := os.Create(strings.TrimSuffix(path, ".json") + "-updated.json")
		if err != nil {
			return nil
		}
		marshalData, err := json.Marshal(changedData)
		if err != nil {
			return nil
		}
		newJson.Write(marshalData)
		defer newJson.Close()
	}
	return nil
}
