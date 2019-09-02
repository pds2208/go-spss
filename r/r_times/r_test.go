package r_times

import (
	"fmt"
	r2 "go-spss/r"
	"log"
	"testing"
)

func Test_times(t *testing.T) {

	t.Logf("Starting test - times")

	r := r2.RFunctions{}
	defer r.Free()

	r.LoadRSource("times.R")

	res, err := Times(5.78, 9.23)

	if err != nil {
		log.Printf("Call to R failed: %s", err)
		panic(err)
	}

	fmt.Printf("Result: %f\n", res)
	t.Logf("Test - add, successful")
}
