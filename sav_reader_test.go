package spss

import (
	"fmt"
	"testing"
)

type SpssFile struct {
	Shiftno float64 `spss:"Shiftno"`
	Serial  float64 `spss:"Serial"`
	Version string  `spss:"Version"`
}

func Test_reader(t *testing.T) {

	var spssFile []*SpssFile

	if err := ReadFromSPSS("testdata/ips1710bv2.sav", &spssFile); err != nil { // Load spssFile from file
		panic(err)
	}

	i := len(spssFile)

	for _, client := range spssFile {
		fmt.Printf("shiftno: %f, serial: %f, version: %s\n", client.Shiftno, client.Serial, client.Version)
	}
	fmt.Printf("Total Items: %d\n", i)

}
