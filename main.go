package main

import (
	"fmt"

	"github.com/mlinhard/exactly-index/esa"
)

func main() {
	esa, err := esa.New(([]byte)("ABRACADABRA"))
	if err != nil {
		fmt.Printf("Error creating enhanced suffix array: %v", err)
		return
	}
	fmt.Print(esa.Print())
}
