package main

// #cgo amd64 CFLAGS: -g
// #cgo LDFLAGS: -lreadstat
// #include <stdlib.h>
// #include "sav_generate.h"
import "C"

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"unsafe"
)

const (
	ReadstatTypeString    = iota
	ReadstatTypeInt8      = iota
	ReadstatTypeInt16     = iota
	ReadstatTypeInt32     = iota
	ReadstatTypeFloat     = iota
	ReadstatTypeDouble    = iota
	ReadstatTypeStringRef = iota
)

type HeaderInfo struct {
	Name    string
	VarType int
}

var header = make([]HeaderInfo, 0)

//export goAddHeaderItem
func goAddHeaderItem(pos C.int, name *C.char, varType C.int, end C.int) {
	if int(end) != 1 { // we are done
		header = append(header, HeaderInfo{C.GoString(name), int(varType)})
	}
}

func Generate(fileName string) int {
	name := C.CString(fileName)
	defer C.free(unsafe.Pointer(name))

	res := C.read_header(name)
	if res != 0 {
		return 1
	}

	return 0
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	input := flag.String("input", "", "input file name")
	output := flag.String("output", "", "output file name")
	structName := flag.String("struct", "Struct", "structure name")
	packageName := flag.String("package", "LFS", "package name")

	flag.Parse()

	f, err := os.Create(*output)
	check(err)

	defer f.Close()

	w := bufio.NewWriter(f)

	Generate(*input)

	fmt.Printf("Found %d Columns\n", len(header))

	if len(header) == 0 {
		fmt.Println("No items found in SPSS file. Is it Empty?")
		os.Exit(1)
	}

	fmt.Fprintf(w, "package %s\n\n", *packageName)
	fmt.Fprintf(w, "type %s struct {\n", *structName)

	for _, j := range header {
		switch j.VarType {
		case ReadstatTypeString:
			fmt.Fprintf(w, "    %s string \t`spss:\"%s\"`\n", j.Name, j.Name)
		case ReadstatTypeInt8:
			fmt.Fprintf(w, "    %s int \t`spss:\"%s\"`\n", j.Name, j.Name)
		case ReadstatTypeInt16:
			fmt.Fprintf(w, "    %s int \t`spss:\"%s\"`\n", j.Name, j.Name)
		case ReadstatTypeInt32:
			fmt.Fprintf(w, "    %s int \t`spss:\"%s\"`\n", j.Name, j.Name)
		case ReadstatTypeFloat:
			fmt.Fprintf(w, "    %s float \t`spss:\"%s\"`\n", j.Name, j.Name)
		case ReadstatTypeDouble:
			fmt.Fprintf(w, "    %s float64 \t`spss:\"%s\"`\n", j.Name, j.Name)
		case ReadstatTypeStringRef:
			panic("String references not supported")
		}
	}

	fmt.Fprintf(w, "}\n")
	w.Flush()
}
