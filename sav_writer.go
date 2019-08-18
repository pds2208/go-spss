package spss

// #cgo windows amd64 CFLAGS: -g -IC:/msys64/mingw64/include
// #cgo windows LDFLAGS: -LC:/msys64/mingw64/libs -lreadstat
// #cgo darwin amd64 CFLAGS: -g
// #cgo darwin LDFLAGS: -lreadstat
// #cgo linux amd64 CFLAGS: -I/usr/local/include -g
// #cgo linux LDFLAGS: -L/usr/local/lib -lreadstat
// #include "sav_writer.h"
// #include <stdlib.h>
import "C"
import "unsafe"

type Header struct {
	SavType ColumnType
	Name    string
	Label   string
}

type DataItem struct {
	Value []interface{}
}

func Export(fileName string, label string, headers []Header, data []DataItem) int {

	numHeaders := len(headers)
	cHeaders := (*[1 << 28]*C.file_header)(C.malloc(C.size_t(C.sizeof_file_header * numHeaders)))
	for i, f := range headers {
		foo := (*C.file_header)(C.malloc(C.size_t(C.sizeof_file_header)))
		(*foo).sav_type = C.int(f.SavType)
		(*foo).name = C.CString(f.Name)
		(*foo).label = C.CString(f.Label)
		cHeaders[i] = foo
	}

	numRows := len(data)
	// DataItem represents a single data item. The number of items is therefore the
	// number of rows multiplied by the number of columns
	cDataItem := (*[1 << 28]*C.data_item)(C.malloc(C.size_t(C.sizeof_data_item * numRows * numHeaders)))

	cnt := 0

	for _, r := range data {

		for j, col := range r.Value {
			dataItem := (*C.data_item)(C.malloc(C.size_t(C.sizeof_data_item)))

			(*dataItem).sav_type = C.int(headers[j].SavType)

			switch headers[j].SavType {

			case ReadstatTypeString:
				if _, ok := col.(string); !ok {
					(*dataItem).string_value = C.CString(col.(string))
					panic("Invalid type, string expected")
				}
				(*dataItem).string_value = C.CString(col.(string))

			case ReadstatTypeInt8:
				if _, ok := col.(int); !ok {
					panic("Invalid type, int8 expected")
				}
				(*dataItem).int_value = C.int(col.(int))

			case ReadstatTypeInt16:
				if _, ok := col.(int); !ok {
					panic("Invalid type, int16 expected")
				}
				(*dataItem).int_value = C.int(col.(int))

			case ReadstatTypeInt32:
				if _, ok := col.(int); !ok {
					panic("Invalid type, int32 expected")
				}
				(*dataItem).int_value = C.int(col.(int))

			case ReadstatTypeFloat:
				if _, ok := col.(float32); !ok {
					panic("Invalid type, float32 expected")
				}
				(*dataItem).float_value = C.float(col.(float32))

			case ReadstatTypeDouble:
				if _, ok := col.(float64); !ok {
					panic("Invalid type, double expected")
				}
				(*dataItem).double_value = C.double(col.(float64))

			case ReadstatTypeStringRef:
				panic("String references not supported")
			}
			cDataItem[cnt] = dataItem
			cnt++
		}
	}

	res := C.save_sav(C.CString(fileName), C.CString(label), &cHeaders[0], C.int(numHeaders), C.int(numRows), &cDataItem[0])

	// Free up C allocated memory
	for i := 0; i < numHeaders; i++ {
		C.free(unsafe.Pointer((*cHeaders[i]).name))
		C.free(unsafe.Pointer((*cHeaders[i]).label))
		C.free(unsafe.Pointer(cHeaders[i]))
	}
	C.free(unsafe.Pointer(cHeaders))

	for i := 0; i < numRows*numHeaders; i++ {
		if int((*cDataItem[i]).sav_type) == ReadstatTypeString.AsInt() {
			C.free(unsafe.Pointer((*cDataItem[i]).string_value))
		}
		C.free(unsafe.Pointer(cDataItem[i]))
	}
	C.free(unsafe.Pointer(cDataItem))

	return int(res)
}
