package spss

// #cgo amd64 CFLAGS: -g
// #cgo LDFLAGS: -lreadstat
// #include "sav_writer.h"
// #include <stdlib.h>
import "C"
import (
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

type SavHeader struct {
	SavType int32
	Name    string
	Label   string
}

type SavData struct {
	SavType int32
	Value   interface{}
}

func ExportSavFile(fileName string, label string, headers []SavHeader, data []SavData) int {

	l := len(headers)
	savHeaders := (*[1 << 28]*C.SavHeader)(C.malloc(C.size_t(C.sizeof_SavHeader * l)))
	for i, f := range headers {
		foo := (*C.SavHeader)(C.malloc(C.size_t(C.sizeof_SavHeader)))
		(*foo).sav_type = C.int(f.SavType)
		(*foo).name = C.CString(f.Name)
		(*foo).label = C.CString(f.Label)
		savHeaders[i] = foo
	}

	d := len(data)
	savData := (*[1 << 28]*C.SavData)(C.malloc(C.size_t(C.sizeof_SavData * d)))
	for i, f := range data {
		foo := (*C.SavData)(C.malloc(C.size_t(C.sizeof_SavData)))
		(*foo).sav_type = C.int(f.SavType)

		switch f.SavType {
		case ReadstatTypeString:
			if _, ok := f.Value.(string); !ok {
				(*foo).string_value = C.CString(f.Value.(string))
				panic("Invalid type, string expected")
			}
			(*foo).string_value = C.CString(f.Value.(string))
		case ReadstatTypeInt8:
			if _, ok := f.Value.(int); !ok {
				panic("Invalid type, int8 expected")
			}
			(*foo).int_value = C.int(f.Value.(int))
		case ReadstatTypeInt16:
			if _, ok := f.Value.(int); !ok {
				panic("Invalid type, int16 expected")
			}
			(*foo).int_value = C.int(f.Value.(int))
		case ReadstatTypeInt32:
			if _, ok := f.Value.(int); !ok {
				panic("Invalid type, int32 expected")
			}
			(*foo).int_value = C.int(f.Value.(int))
		case ReadstatTypeFloat:
			if _, ok := f.Value.(float32); !ok {
				panic("Invalid type, float32 expected")
			}
			(*foo).float_value = C.float(f.Value.(float32))
		case ReadstatTypeDouble:
			if _, ok := f.Value.(float64); !ok {
				panic("Invalid type, double expected")
			}
			(*foo).double_value = C.double(f.Value.(float64))
		case ReadstatTypeStringRef:
			panic("String references not supported")
		}

		savData[i] = foo
	}

	res := C.save_sav(C.CString(fileName), C.CString(label), &savHeaders[0], C.int(l), C.int(d), &savData[0])

	for i := 0; i < l; i++ {
		C.free(unsafe.Pointer((*savHeaders[i]).name))
		C.free(unsafe.Pointer((*savHeaders[i]).label))
		C.free(unsafe.Pointer(savHeaders[i]))
	}
	C.free(unsafe.Pointer(savHeaders))

	for i := 0; i < d; i++ {
		if (*savData[i]).sav_type == ReadstatTypeString {
			C.free(unsafe.Pointer((*savData[i]).string_value))
		}
		C.free(unsafe.Pointer(savData[i]))
	}
	C.free(unsafe.Pointer(savData))

	return int(res)
}
