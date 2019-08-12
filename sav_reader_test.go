package spss

import (
	"testing"
)

type SpssFile struct {
	Shiftno float64 `spss:"Shiftno"`
	Serial  float64 `spss:"Serial"`
	Version string  `spss:"Version"`
}

func Test_reader(t *testing.T) {

	var spssFile []*SpssFile

	if err := ReadFromSPSSFile("testdata/ips1710bv2.sav", &spssFile); err != nil { // Load spssFile from file
		panic(err)
	}

	t.Logf("Starting test - reader")

	i := len(spssFile)

	t.Logf("Total Items: %d\n", i)

	t.Logf("Test finished - reader")

}

type Mydataset struct {
	Shiftno float64 `spss:"Shiftno"`
	Serial  float64 `spss:"Serial"`
	Version string  `spss:"Version"`
}

func Test_dataset(t *testing.T) {

	var spssFile []Mydataset

	if err := ReadFromSPSSFile("testdata/ips1710bv2.sav", &spssFile); err != nil { // Load spssFile from file
		panic(err)
	}

	t.Logf("Starting test - reader")

	i := len(spssFile)
	t.Logf("Total Items: %d\n", i)

	//t.Logf("Mean: %f\n", a)
	t.Logf("Test finished - reader")

}
