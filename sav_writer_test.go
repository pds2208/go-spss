package spss

import (
	"fmt"
	"testing"
)

type SpssWriteFile struct {
	Shiftno float64 `spss:"Shiftno"`
	Serial  float64 `spss:"Serial"`
	Version string  `spss:"Version"`
}

func createHeader(wr []SpssWriteFile) []Header {
	header := make([]Header, 0)

	header = append(header, Header{ReadstatTypeDouble, "Shiftno", "Shiftno Label"})
	header = append(header, Header{ReadstatTypeDouble, "Serial", "Serial Label"})
	header = append(header, Header{ReadstatTypeString, "Version", "Version Label"})

	return header
}

func createData(wr []SpssWriteFile) []DataItem {

	data := make([]DataItem, 0)
	for _, j := range wr {
		data = append(data, DataItem{[]interface{}{j.Shiftno, j.Serial, j.Version}})
	}

	return data
}

func Test_writer(t *testing.T) {

	wr := []SpssWriteFile{
		{1.0, 123456.00, "v1"},
		{2.0, 789012.00, "v2"},
		{3.0, 789888.00, "v2"},
	}

	header := createHeader(wr)
	data := createData(wr)

	val := ExportSavFile("/Users/paul/Desktop/test_output.sav", "Test SAV from GO",
		header, data)

	fmt.Println("Finished, return value: ", val)

}
