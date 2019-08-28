package r

import (
	"log"
	"testing"
)

func Test_add(t *testing.T) {

	t.Logf("Starting test - add")

	r := rFunctions{}
	defer r.free()

	arg := []int{1, 2, 3, 4, 5}
	err := r.AddArray(arg)

	if err != nil {
		log.Printf("Call to R failed: %s", err)
		panic(err)
	}

	t.Logf("Test - add, successful")
}
