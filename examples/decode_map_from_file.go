package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ioannuwu/vcard"
)

func decodeSingleVCardFromFileAndPrint() {

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
	fmt.Printf("data:\n%s\n", data)

	// Unmarshal this byte slice into a map according to default schema
	m := make(map[string]string)
	vcard.Unmarshal(data, &m)

	// Print results quoted
	fmt.Printf("map:\n")
	for k, v := range m {
		fmt.Printf("%q -> %q\n", k, v)
	}
}

func decodeMultipleVCardsFromFileAndPrint() {

	// This file contains multiple vCards
	filePath := "assets/vCards.vcf"

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

	// Unmarshal this byte slice into a slice of maps according to default schema
	slice := make([]map[string]string, 0)
	err = vcard.Unmarshal(data, &slice)

	// Print results quoted
	fmt.Printf("map:\n")

	for idx, m := range slice {
		fmt.Printf("--- Map #%v ---\n", idx)
		for k, v := range m {
			fmt.Printf("%q -> %q\n", k, v)
		}
	}
}
