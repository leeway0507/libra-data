package main

import (
	"fmt"
	"libraData/collect"
	"log"
	"strconv"
	"strings"
)

func main() {
	// 양천 111015 강서 111005
	scraps := []string{
		"111003,2023-01-01,2023-12-31",
		"111004,2023-01-01,2023-12-31",
		"111006,2023-01-01,2023-12-31",
		"111003,2024-01-01,2024-12-01",
		"111004,2024-01-01,2024-12-01",
		"111006,2024-01-01,2024-12-01",
	}

	for _, scrap := range scraps {
		fmt.Printf("start %v", scrap)
		a := strings.Split(scrap, ",")
		libCode, err := strconv.Atoi(a[0])
		if err != nil {
			log.Fatalln(err)
		}
		collect.GetAllBooksFromLib(libCode, a[1], a[2])
		fmt.Printf("end %v", scrap)
	}
}
