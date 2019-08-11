package spss

// #cgo amd64 CFLAGS: -g -IC:/msys64/mingw64/include
// #cgo LDFLAGS: -LC:/msys64/mingw64/libs -lreadstat
// #include <stdlib.h>
// #include "sav_reader.h"
import "C"

import (
	"bytes"
	"errors"
	"strings"
	"unsafe"
)

var lines = make([]bytes.Buffer, 0)
var headerItems bytes.Buffer

const (
	EOL   = "\n"
	COMMA = ","
)

//export goAddLine
func goAddLine(str *C.char) {
	lines = append(lines, *bytes.NewBufferString(C.GoString(str)))
}

//export goAddHeaderLine
func goAddHeaderLine(pos C.int, name *C.char, varType C.int, end C.int) {
	if int(end) == 1 { // we are done
		headerItems.WriteString(EOL)
		lines = append(lines, headerItems)
	} else {
		headerItems.WriteString(C.GoString(name))
		headerItems.WriteString(COMMA)
	}
}

func GetHeader(fileName string) int {
	name := C.CString(fileName)
	defer C.free(unsafe.Pointer(name))

	res := C.parse_sav(name)
	if res != 0 {
		return 1
	}

	return 0
}

func Import(fileName string) ([][]string, error) {
	name := C.CString(fileName)
	defer C.free(unsafe.Pointer(name))

	res := C.parse_sav(name)
	if res != 0 {
		return nil, errors.New("read from SPSS file failed")
	}
	str := make([][]string, 0)

	for _, l := range lines {
		s := strings.Split(l.String(), TagSeparator)
		str = append(str, s)
	}

	return str, nil
}
