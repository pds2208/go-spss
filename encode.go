package spss

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

type InputType interface{}

type SPSSWriter interface {
	Write() func(rows interface{}) error
}

type BufferOutput struct {
	inputType string
}

// Example implementation
func (b BufferOutput) Write() func(rows interface{}) error {
	return func(rows interface{}) error {
		return nil
	}
}

type FileOutput struct {
	inputType string
}

func (f FileOutput) Write() func(rows interface{}) error {

	return func(rows interface{}) error {

		inValue, inType := getConcreteReflectValueAndType(rows) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
		if err := ensureInType(inType); err != nil {
			return err
		}

		inInnerWasPointer, inInnerType := getConcreteContainerInnerType(inType) // Get the concrete inner type (not pointer) (Container<"?">)
		if err := ensureInInnerType(inInnerType); err != nil {
			return err
		}

		inInnerStructInfo := getStructInfo(inInnerType) // Get the inner struct info to get SPSS annotations
		header := make([]Header, 0)
		data := make([]DataItem, 0)

		for _, fieldInfo := range inInnerStructInfo.Fields { // Used to write metadata rows SPSS

			var spssType int32 = 0

			switch fieldInfo.FieldType {
			case reflect.String:
				spssType = ReadstatTypeString
			case reflect.Int8, reflect.Uint8:
				spssType = ReadstatTypeInt8
			case reflect.Int, reflect.Int32, reflect.Uint32:
				spssType = ReadstatTypeInt32
			case reflect.Float32:
				spssType = ReadstatTypeFloat
			case reflect.Float64:
				spssType = ReadstatTypeDouble
			default:
				return fmt.Errorf("cannot convert type for struct variable %s into SPSS type", fieldInfo.keys[0])
			}

			header = append(header, Header{spssType, fieldInfo.keys[0], fieldInfo.keys[0] + ""})
		}

		if inValue.Kind() != reflect.Slice {
			panic("You need to pass a slice of interface{} to save to an SPSS file")
		}

		inLen := inValue.Len()
		for i := 0; i < inLen; i++ { // Iterate over container rows
			dataItem := make([]interface{}, 0)
			for j, fieldInfo := range inInnerStructInfo.Fields {
				header[j].Label = ""
				inInnerFieldValue, err := getInnerField(inValue.Index(i), inInnerWasPointer, fieldInfo.IndexChain) // Get the correct field header <-> position
				if err != nil {
					return err
				}
				// convert to correct type
				var spssType interface{}

				switch fieldInfo.FieldType {
				case reflect.String:
					spssType = inInnerFieldValue
				case reflect.Int8, reflect.Uint8:
					spssType, _ = strconv.Atoi(inInnerFieldValue)
				case reflect.Int, reflect.Int32, reflect.Uint32:
					spssType, _ = strconv.Atoi(inInnerFieldValue)
				case reflect.Float32:
					spssType, _ = strconv.ParseFloat(inInnerFieldValue, 32)
				case reflect.Float64:
					spssType, _ = strconv.ParseFloat(inInnerFieldValue, 64)
				default:
					return fmt.Errorf("cannot convert value for struct variable %s into SPSS type", fieldInfo.keys[0])
				}

				dataItem = append(dataItem, spssType)
			}
			data = append(data, DataItem{dataItem})

		}

		val := Export(f.inputType, "Test SAV from GO", header, data)
		log.Printf("Finished writing to: %s, return value: %d", f.inputType, val)

		return nil
	}
}

// Check if the inType is an array or a slice
func ensureInType(outType reflect.Type) error {
	switch outType.Kind() {
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		return nil
	}
	return fmt.Errorf("cannot use " + outType.String() + ", only slice or array supported")
}

// Check if the inInnerType is of type struct
func ensureInInnerType(outInnerType reflect.Type) error {
	switch outInnerType.Kind() {
	case reflect.Struct:
		return nil
	}
	return fmt.Errorf("cannot use " + outInnerType.String() + ", only struct supported")
}

func getInnerField(outInner reflect.Value, outInnerWasPointer bool, index []int) (string, error) {
	oi := outInner
	if outInnerWasPointer {
		if oi.IsNil() {
			return "", nil
		}
		oi = outInner.Elem()
	}
	// because pointers can be nil need to recurse one index at a time and perform nil check
	if len(index) > 1 {
		nextField := oi.Field(index[0])
		return getInnerField(nextField, nextField.Kind() == reflect.Ptr, index[1:])
	}
	return getFieldAsString(oi.FieldByIndex(index))
}
