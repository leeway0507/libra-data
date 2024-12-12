package main

import "libraData/collect"

func Main() {
	libCode, startDate, endDate := 111015, "2024-11-01", "2024-11-30"
	collect.GetBookItemsAll(libCode, startDate, endDate)
}
