package spss

// #cgo amd64 CFLAGS: -g
// #cgo LDFLAGS: -lreadstat
// #include <stdlib.h>
// #include "sav_reader.h"
import "C"

import (
	"bytes"
	"io"
	"strings"
	"unsafe"
)

var lines = make([]bytes.Buffer, 0)
var headerItems bytes.Buffer

//export goAddLine
func goAddLine(str *C.char) {
	lines = append(lines, *bytes.NewBufferString(C.GoString(str)))
}

//export goAddHeaderLine
func goAddHeaderLine(pos C.int, name *C.char, varType C.int, end C.int) {
	if int(end) == 1 { // we are done
		headerItems.WriteString("\n")
		lines = append(lines, headerItems)
	} else {
		headerItems.WriteString(C.GoString(name))
		headerItems.WriteString(",")
	}
}

func Import(fileName string) int {
	name := C.CString(fileName)
	defer C.free(unsafe.Pointer(name))

	res := C.parse_sav(name)
	if res != 0 {
		return 1
	}

	return 0
}

type Reader struct {
	fileName    string
	currentLine int
	eof         bool
}

func NewReader(f string) *Reader {
	return &Reader{fileName: f, currentLine: 0, eof: false}
}

func (r *Reader) Read() ([]string, error) {

	if len(lines) == r.currentLine-1 {
		r.eof = true
	}

	if r.eof {
		return nil, io.EOF
	}

	str := lines[r.currentLine].String()
	r.currentLine++
	return strings.Split(str, TagSeparator), nil
}

func (r *Reader) ReadAll() ([][]string, error) {
	if r.eof {
		return nil, io.EOF
	}

	r.eof = true

	str := make([][]string, 0)

	for _, l := range lines {
		s := strings.Split(l.String(), TagSeparator)
		str = append(str, s)
	}

	return str, nil
}
