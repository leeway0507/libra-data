package main

import "libraData/collection"

func main() {
	// 양천 111015 강서 111005
	libCode, startDate, endDate := 111005, "2024-10-01", "2024-11-01"
	collection.GetAllBooksFromLib(libCode, startDate, endDate)
}
