package spss

// #cgo amd64 CFLAGS: -g
// #cgo LDFLAGS: -lreadstat
// #include <stdlib.h>
// #include "import_sav.h"
import "C"

import (
    "bufio"
    "fmt"
    "io"
    "unsafe"
)

type Line string

var lines [] Line

type headerLine struct {
    vType    int
    position int
}

var headerItems = make(map[string]headerLine)

//export goAddLine
func goAddLine(str *C.char) { lines = append(lines, C.GoString(str)) }

//export goAddHeaderLine
func goAddHeaderLine(pos C.int, name *C.char, varType C.int, end C.int) {
    if int(end) == 1 { // we are done
        fmt.Printf("Header %v", headerItems)
    } else {
        headerItems[C.GoString(name)] = headerLine{int(varType), int(pos)}
    }
}

func Import(fileName ImportFile) int {
    name := C.CString(fileName)
    defer C.free(unsafe.Pointer(name))

    res := C.parse_sav(name)
    if res != 0 {
        return 1
    }

    return 0
}

type Reader struct {
    r *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
    return &Reader{r: bufio.NewReader(r)}
}

func (r *Reader) Read() ([]string, error) {

}

func (r *Reader) ReadAll() ([][]string, error) {

}
