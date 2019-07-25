package spss

// #cgo amd64 CFLAGS: -g
// #cgo LDFLAGS: -lreadstat9
// #include "sav_writer.h"
import "C"
import "unsafe"

const (
	READSTAT_TYPE_STRING     = iota
	READSTAT_TYPE_INT8       = iota
	READSTAT_TYPE_INT16      = iota
	READSTAT_TYPE_INT32      = iota
	READSTAT_TYPE_FLOAT      = iota
	READSTAT_TYPE_DOUBLE     = iota
	READSTAT_TYPE_STRING_REF = iota
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
		(*foo).value = unsafe.Pointer(&f.Value)
		savData[i] = foo
	}

	res := C.save_sav(C.CString(fileName), C.CString(label), &savHeaders[0],
		C.int(l), C.int(d), &savData[0])

	for i := 0; i < l; i++ {
		C.free(unsafe.Pointer(savHeaders[i]))
	}
	C.free(unsafe.Pointer(savHeaders))

	for i := 0; i < d; i++ {
		C.free(unsafe.Pointer(savData[i]))
	}
	C.free(unsafe.Pointer(savData))

	return res
}
