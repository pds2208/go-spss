package spss

import (
	"fmt"
	"testing"
)

func Test_writer(t *testing.T) {

	header := []SavHeader{
		{READSTAT_TYPE_STRING, "ColumnOne", "ColumnOne Label"},
		{READSTAT_TYPE_DOUBLE, "ColumnTwo", "ColumnTwo Label"},
	}

	vs := "This is item one"
	vd := 22.99

	data := []SavData{
		{READSTAT_TYPE_STRING, &vs},
		{READSTAT_TYPE_DOUBLE, &vd},
	}

	val := ExportSavFile("/Users/paul/Desktop/test_output.sav", "Test SAV from GO",
		header, data)

	fmt.Println("A finished", val)

}
