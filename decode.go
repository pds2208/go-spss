package spss

import (
    "encoding/csv"
    "errors"
    "fmt"
    "io"
    "reflect"
)

// Decoder .
type Decoder interface {
    getSPSSRows() ([][]string, error)
}

// SimpleDecoder .
type SimpleDecoder interface {
    getSPSSRow() ([]string, error)
    getSPSSRows() ([][]string, error)
}

type SPSSReader interface {
    Read() ([]string, error)
    ReadAll() ([][]string, error)
}

type spssDecoder struct {
    SPSSReader
}

func newSimpleDecoderFromReader(r io.Reader) SimpleDecoder {
    return spssDecoder{getSPSSReader(r)}
}

// NewSimpleDecoderFromSPSSReader creates a SimpleDecoder, which may be passed
// to the UnmarshalDecoder* family of functions, from a SPSS reader.
func NewSimpleDecoderFromSPSSReader(r SPSSReader) SimpleDecoder {
    return spssDecoder{r}
}

func (c spssDecoder) getSPSSRows() ([][]string, error) {
    return c.ReadAll()
}

func (c spssDecoder) getSPSSRow() ([]string, error) {
    return c.Read()
}

func mismatchStructFields(structInfo []fieldInfo, headers []string) []string {
    missing := make([]string, 0)
    if len(structInfo) == 0 {
        return missing
    }

    headerMap := make(map[string]struct{}, len(headers))
    for idx := range headers {
        headerMap[headers[idx]] = struct{}{}
    }

    for _, info := range structInfo {
        found := false
        for _, key := range info.keys {
            if _, ok := headerMap[key]; ok {
                found = true
                break
            }
        }
        if !found {
            missing = append(missing, info.keys...)
        }
    }
    return missing
}

func mismatchHeaderFields(structInfo []fieldInfo, headers []string) []string {
    missing := make([]string, 0)
    if len(headers) == 0 {
        return missing
    }

    keyMap := make(map[string]struct{})
    for _, info := range structInfo {
        for _, key := range info.keys {
            keyMap[key] = struct{}{}
        }
    }

    for _, header := range headers {
        if _, ok := keyMap[header]; !ok {
            missing = append(missing, header)
        }
    }
    return missing
}

func maybeMissingStructFields(structInfo []fieldInfo, headers []string) error {
    missing := mismatchStructFields(structInfo, headers)
    if len(missing) != 0 {
        return fmt.Errorf("found unmatched struct field with tags %v", missing)
    }
    return nil
}

// Check that no header name is repeated twice
func maybeDoubleHeaderNames(headers []string) error {
    headerMap := make(map[string]bool, len(headers))
    for _, v := range headers {
        if _, ok := headerMap[v]; ok {
            return fmt.Errorf("repeated header name: %v", v)
        }
        headerMap[v] = true
    }
    return nil
}

func readTo(decoder Decoder, out interface{}) error {
    outValue, outType := getConcreteReflectValueAndType(out) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
    if err := ensureOutType(outType); err != nil {
        return err
    }
    outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
    if err := ensureOutInnerType(outInnerType); err != nil {
        return err
    }
    spssRows, err := decoder.getSPSSRows()
    if err != nil {
        return err
    }
    if len(spssRows) == 0 {
        return errors.New("empty spss file given")
    }
    if err := ensureOutCapacity(&outValue, len(spssRows)); err != nil { // Ensure the container is big enough to hold the SPSS content
        return err
    }
    outInnerStructInfo := getStructInfo(outInnerType) // Get the inner struct info to get SPSS annotations
    if len(outInnerStructInfo.Fields) == 0 {
        return errors.New("no spss struct tags found")
    }

    headers := spssRows[0]
    body := spssRows[1:]

    spssHeadersLabels := make(map[int]*fieldInfo, len(outInnerStructInfo.Fields)) // Used to store the correspondance header <-> position in CSV

    headerCount := map[string]int{}
    for i, csvColumnHeader := range headers {
        curHeaderCount := headerCount[csvColumnHeader]
        if fieldInfo := getCSVFieldPosition(csvColumnHeader, outInnerStructInfo, curHeaderCount); fieldInfo != nil {
            spssHeadersLabels[i] = fieldInfo

        }
    }

    if FailIfUnmatchedStructTags {
        if err := maybeMissingStructFields(outInnerStructInfo.Fields, headers); err != nil {
            return err
        }
    }
    if FailIfDoubleHeaderNames {
        if err := maybeDoubleHeaderNames(headers); err != nil {
            return err
        }
    }

    for i, csvRow := range body {
        outInner := createNewOutInner(outInnerWasPointer, outInnerType)
        for j, csvColumnContent := range csvRow {
            if fieldInfo, ok := spssHeadersLabels[j]; ok { // Position found accordingly to header name
                if err := setInnerField(&outInner, outInnerWasPointer, fieldInfo.IndexChain, csvColumnContent, fieldInfo.omitEmpty); err != nil { // Set field of struct
                    return &csv.ParseError{
                        Line:   i + 2, //add 2 to account for the header & 0-indexing of arrays
                        Column: j + 1,
                        Err:    err,
                    }
                }
            }
        }
        outValue.Index(i).Set(outInner)
    }
    return nil
}

func readEach(decoder SimpleDecoder, c interface{}) error {
    headers, err := decoder.getSPSSRow()
    if err != nil {
        return err
    }
    outValue, outType := getConcreteReflectValueAndType(c) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
    if err := ensureOutType(outType); err != nil {
        return err
    }
    defer outValue.Close()
    outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
    if err := ensureOutInnerType(outInnerType); err != nil {
        return err
    }
    outInnerStructInfo := getStructInfo(outInnerType) // Get the inner struct info to get CSV annotations
    if len(outInnerStructInfo.Fields) == 0 {
        return errors.New("no csv struct tags found")
    }
    csvHeadersLabels := make(map[int]*fieldInfo, len(outInnerStructInfo.Fields)) // Used to store the correspondance header <-> position in CSV
    headerCount := map[string]int{}
    for i, csvColumnHeader := range headers {
        curHeaderCount := headerCount[csvColumnHeader]
        if fieldInfo := getCSVFieldPosition(csvColumnHeader, outInnerStructInfo, curHeaderCount); fieldInfo != nil {
            csvHeadersLabels[i] = fieldInfo
            if ShouldAlignDuplicateHeadersWithStructFieldOrder {
                curHeaderCount++
                headerCount[csvColumnHeader] = curHeaderCount
            }
        }
    }
    if err := maybeMissingStructFields(outInnerStructInfo.Fields, headers); err != nil {
        if FailIfUnmatchedStructTags {
            return err
        }
    }
    if FailIfDoubleHeaderNames {
        if err := maybeDoubleHeaderNames(headers); err != nil {
            return err
        }
    }
    i := 0
    for {
        line, err := decoder.getSPSSRow()
        if err == io.EOF {
            break
        } else if err != nil {
            return err
        }
        outInner := createNewOutInner(outInnerWasPointer, outInnerType)
        for j, csvColumnContent := range line {
            if fieldInfo, ok := csvHeadersLabels[j]; ok { // Position found accordingly to header name
                if err := setInnerField(&outInner, outInnerWasPointer, fieldInfo.IndexChain, csvColumnContent, fieldInfo.omitEmpty); err != nil { // Set field of struct
                    return &csv.ParseError{
                        Line:   i + 2, //add 2 to account for the header & 0-indexing of arrays
                        Column: j + 1,
                        Err:    err,
                    }
                }
            }
        }
        outValue.Send(outInner)
        i++
    }
    return nil
}

func readToWithoutHeaders(decoder Decoder, out interface{}) error {
    outValue, outType := getConcreteReflectValueAndType(out) // Get the concrete type (not pointer) (Slice<?> or Array<?>)
    if err := ensureOutType(outType); err != nil {
        return err
    }
    outInnerWasPointer, outInnerType := getConcreteContainerInnerType(outType) // Get the concrete inner type (not pointer) (Container<"?">)
    if err := ensureOutInnerType(outInnerType); err != nil {
        return err
    }
    csvRows, err := decoder.getSPSSRows()
    if err != nil {
        return err
    }
    if len(csvRows) == 0 {
        return errors.New("empty csv file given")
    }
    if err := ensureOutCapacity(&outValue, len(csvRows)+1); err != nil { // Ensure the container is big enough to hold the CSV content
        return err
    }
    outInnerStructInfo := getStructInfo(outInnerType) // Get the inner struct info to get CSV annotations
    if len(outInnerStructInfo.Fields) == 0 {
        return errors.New("no csv struct tags found")
    }

    for i, csvRow := range csvRows {
        outInner := createNewOutInner(outInnerWasPointer, outInnerType)
        for j, csvColumnContent := range csvRow {
            fieldInfo := outInnerStructInfo.Fields[j]
            if err := setInnerField(&outInner, outInnerWasPointer, fieldInfo.IndexChain, csvColumnContent, fieldInfo.omitEmpty); err != nil { // Set field of struct
                return &csv.ParseError{
                    Line:   i + 1,
                    Column: j + 1,
                    Err:    err,
                }
            }
        }
        outValue.Index(i).Set(outInner)
    }

    return nil
}

// Check if the outType is an array or a slice
func ensureOutType(outType reflect.Type) error {
    switch outType.Kind() {
    case reflect.Slice:
        fallthrough
    case reflect.Chan:
        fallthrough
    case reflect.Array:
        return nil
    }
    return fmt.Errorf("cannot use " + outType.String() + ", only slice or array supported")
}

// Check if the outInnerType is of type struct
func ensureOutInnerType(outInnerType reflect.Type) error {
    switch outInnerType.Kind() {
    case reflect.Struct:
        return nil
    }
    return fmt.Errorf("cannot use " + outInnerType.String() + ", only struct supported")
}

func ensureOutCapacity(out *reflect.Value, csvLen int) error {
    switch out.Kind() {
    case reflect.Array:
        if out.Len() < csvLen-1 { // Array is not big enough to hold the CSV content (arrays are not addressable)
            return fmt.Errorf("array capacity problem: cannot store %d %s in %s", csvLen-1, out.Type().Elem().String(), out.Type().String())
        }
    case reflect.Slice:
        if !out.CanAddr() && out.Len() < csvLen-1 { // Slice is not big enough tho hold the CSV content and is not addressable
            return fmt.Errorf("slice capacity problem and is not addressable (did you forget &?)")
        } else if out.CanAddr() && out.Len() < csvLen-1 {
            out.Set(reflect.MakeSlice(out.Type(), csvLen-1, csvLen-1)) // Slice is not big enough, so grows it
        }
    }
    return nil
}

func getCSVFieldPosition(key string, structInfo *structInfo, curHeaderCount int) *fieldInfo {
    matchedFieldCount := 0
    for _, field := range structInfo.Fields {
        if field.matchesKey(key) {
            if matchedFieldCount >= curHeaderCount {
                return &field
            }
            matchedFieldCount++
        }
    }
    return nil
}

func createNewOutInner(outInnerWasPointer bool, outInnerType reflect.Type) reflect.Value {
    if outInnerWasPointer {
        return reflect.New(outInnerType)
    }
    return reflect.New(outInnerType).Elem()
}

func setInnerField(outInner *reflect.Value, outInnerWasPointer bool, index []int, value string, omitEmpty bool) error {
    oi := *outInner
    if outInnerWasPointer {
        // initialize nil pointer
        if oi.IsNil() {
            setField(oi, "", omitEmpty)
        }
        oi = outInner.Elem()
    }
    // because pointers can be nil need to recurse one index at a time and perform nil check
    if len(index) > 1 {
        nextField := oi.Field(index[0])
        return setInnerField(&nextField, nextField.Kind() == reflect.Ptr, index[1:], value, omitEmpty)
    }
    return setField(oi.FieldByIndex(index), value, omitEmpty)
}
