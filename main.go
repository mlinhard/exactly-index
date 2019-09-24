package main

import (
	"fmt"

	"github.com/mlinhard/exactly-index/esa"
)

func main() {

	bytes := ([]byte)("ABRACADABRA")
	esa1, err := esa.New(bytes)
	if err != nil {
		fmt.Printf("Error creating enhanced suffix array: %v", err)
		return
	}
	var suffixes [][]byte
	for i := range bytes {
		suffixes = append(suffixes, bytes[i:])
	}
	suffixes = esa.SortBAs(suffixes)
	for i := range suffixes {
		expected := string(suffixes[i])
		computed := string(bytes[esa1.SA[i]:])
		res := "OK"
		if expected != computed {
			res = "FAIL: " + computed
		}
		fmt.Printf("%2v %v %v\n", i, expected, res)
	}
	//fmt.Print(esa1.Print())

}
