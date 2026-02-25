package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ioannuwu/vcard"
)

func decodeFromFileThenEncodeBack() {

	// This file contains 1 vCard
	filePath := "assets/vCard.vcf"

	// Open file and read it into a byte slice
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	// Print original vCard data from file
	fmt.Printf("original:\n%s\n", data)

	m := make(map[string]string)
	vcard.Unmarshal(data, &m)

	b, _ := vcard.Marshal(m)

	fmt.Printf("normalized:\n%s\n", b)

	// Lets marshal and unmarshal again

	mm := make(map[string]string)
	vcard.Unmarshal(b, &mm)

	bm, _ := vcard.Marshal(mm)

	fmt.Printf("normalized 2nd time:\n%s\n", bm)
}