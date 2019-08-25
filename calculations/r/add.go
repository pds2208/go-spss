package r

// #cgo windows amd64 CFLAGS: -g -IC:/Users/pauld/scoop/apps/r/current/include
// #cgo windows LDFLAGS: -LC:/Users/pauld/scoop/apps/r/current/bin/curr_arch -lR -lRblas
// #cgo darwin amd64 CFLAGS: -g -I/Library/Frameworks/R.framework/Resources/include
// #cgo darwin LDFLAGS: -L/Library/Frameworks/R.framework/Resources/lib -lR -lRblas
// #cgo linux amd64 CFLAGS: -I/usr/share/R/include -g
// #cgo linux LDFLAGS: -L/usr/lib/R -lR
// #include <stdlib.h>
// #include "r_integration.h"
import "C"
import (
	"errors"
	"unsafe"
)

type rFunctions struct{}

func init() {
	C.initialise()
}

func (r rFunctions) free() {
	C.free_r()
}

func (r rFunctions) AddArray(arg []int) error {
	num := C.int(len(arg))
	var res = C.r_add_array(num, (*C.int)(unsafe.Pointer(&arg[0])))

	if res != 0 {
		return errors.New("R call failed")
	}
	return nil
}
