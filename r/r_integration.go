package r

// #cgo windows amd64 CFLAGS: -g -IC:/Users/pauld/scoop/apps/r/current/include
// #cgo windows LDFLAGS: -LC:/Users/pauld/scoop/apps/r/current/bin/x64 -lR -lRblas
// #cgo darwin amd64 CFLAGS: -g -I/Library/Frameworks/R.framework/Resources/include
// #cgo darwin LDFLAGS: -L/Library/Frameworks/R.framework/Resources/lib -lR -lRblas
// #cgo linux amd64 CFLAGS: -I/usr/share/R/include -g
// #cgo linux LDFLAGS: -L/usr/lib/R -lR
// #include <stdlib.h>
// #include "r_integration.h"
import "C"
import (
	"unsafe"
)

type RFunctions struct{}

func init() {
	C.initialise()
}

func (r RFunctions) Free() {
	C.free_r()
}

func (r RFunctions) LoadRSource(source string) {
	cs := C.CString(source)
	defer C.free(unsafe.Pointer(cs))
	C.load_r_source(cs)
}
