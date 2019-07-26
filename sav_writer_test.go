package spss

import (
	"fmt"
	"testing"
)

func Test_writer(t *testing.T) {

	header := []SavHeader{
		{ReadstatTypeString, "ColumnOne", "ColumnOne Label"},
		{ReadstatTypeDouble, "ColumnTwo", "ColumnTwo Label"},
	}

	data := []SavData{
		{ReadstatTypeString, "This is item one"},
		{ReadstatTypeDouble, 22.99},
		{ReadstatTypeString, "This is item Two"},
		{ReadstatTypeDouble, 222.99},
	}

	val := ExportSavFile("/Users/paul/Desktop/test_output.sav", "Test SAV from GO",
		header, data)

	fmt.Println("Finished, return value: ", val)

}
