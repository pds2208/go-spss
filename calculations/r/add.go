package r

// #cgo windows amd64 CFLAGS: -O3 -IC:/msys64/mingw64/include
// #cgo windows LDFLAGS: -LC:/msys64/mingw64/libs -lreadstat
// #cgo darwin amd64 CFLAGS: -g -I/Library/Frameworks/R.framework/Resources/include
// #cgo darwin LDFLAGS: -L/Library/Frameworks/R.framework/Resources/lib -lR -lRblas
// #cgo linux amd64 CFLAGS: -I/usr/local/include -g
// #cgo linux LDFLAGS: -L/usr/local/lib -lR -lRblas
// #include <stdlib.h>
// #include "r_integration.h"
import "C"
import (
	"errors"
	"unsafe"
)

type rFunctions struct{}

//
//func (r rFunctions) init() {
//    C.initialise()
//}

func (r rFunctions) free() {
	C.free_r()
}

func (r rFunctions) AddArray(arg []int) error {
	num := C.int(len(arg))
	//C.initialise()
	var res = C.r_add_array(num, (*C.int)(unsafe.Pointer(&arg[0])), 5)

	if res != 0 {
		return errors.New("R call failed")
	}
	return nil
}
