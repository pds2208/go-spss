package spss

import (
	"testing"
)

type SpssWriteFile struct {
	Shiftno float64 `spss:"Shiftno"`
	Serial  float64 `spss:"Serial"`
	Version string  `spss:"Version"`
}

func Test_writer(t *testing.T) {

	wr := []SpssWriteFile{
		{1.0, 123456.00, "v1"},
		{2.0, 789012.00, "v2"},
		{3.0, 789888.00, "v2"},
	}

	t.Logf("Starting test - writer")
	err := WriteToSPSS("/Users/paul/Desktop/test_output.sav", &wr)
	if err != nil {
		panic(err)
	}
	t.Logf("Test finished - writer")
}