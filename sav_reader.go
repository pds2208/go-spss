package spss

// #cgo windows amd64 CFLAGS: -g -IC:/msys64/mingw64/include
// #cgo darwin amd64 CFLAGS: -g
// #cgo linux amd64 CFLAGS: -g
// #cgo windows LDFLAGS: -LC:/msys64/mingw64/libs -lreadstat
// #cgo darwin LDFLAGS: -lreadstat
// #cgo linux LDFLAGS: -lreadstat
// #include <stdlib.h>
// #include "sav_reader.h"
import "C"

import (
	"errors"
	"strings"
	"unsafe"
)

func Import(fileName string) ([][]string, error) {
	name := C.CString(fileName)
	defer C.free(unsafe.Pointer(name))

	var res = C.parse_sav(name)
	if res == nil {
		return nil, errors.New("read from SPSS file failed")
	}

	var str [][]string

	defer func() {
		if res == nil {
			return
		}
		if res.buffer != nil {
			C.free(unsafe.Pointer(res.buffer))
		}
		if res.header != nil {
			C.free(unsafe.Pointer(res.header))
		}
		if res.data != nil {
			C.free(unsafe.Pointer(res.data))
		}
		C.free(unsafe.Pointer(res))
	}()

	v := C.struct_Data(*res)

	header := []string{C.GoString(v.header)}
	for _, l := range header {
		s := strings.Split(l, TagSeparator)
		str = append(str, s)
	}

	data := strings.Split(C.GoString(v.data), EOL)

	for _, l := range data {
		s := strings.Split(l, TagSeparator)
		str = append(str, s)
	}

	return str, nil
}
