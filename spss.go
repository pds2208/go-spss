package spss

import (
    "bytes"
    "io"
    "os"
    "strings"
)

// Use of this source code is governed by a MIT license
// The license can be found in the LICENSE file.

// The GoSPSS package aims to provide  SPSS serialization and deserialization

// FailIfUnmatchedStructTags indicates whether it is considered an error when there is an unmatched
// struct tag.
var FailIfUnmatchedStructTags = false

// FailIfDoubleHeaderNames indicates whether it is considered an error when a header name is repeated
// in the csv header.
var FailIfDoubleHeaderNames = false

// ShouldAlignDuplicateHeadersWithStructFieldOrder indicates whether we should align duplicate CSV
// headers per their alignment in the struct definition.
var ShouldAlignDuplicateHeadersWithStructFieldOrder = false

var TagSeparator = ","

var spssReader = DefaultSPSSReader

func DefaultSPSSReader(in io.Reader) SPSSReader {
    return NewReader(in)
}

func SetSPSSReader(reader func(io.Reader) SPSSReader) {
    spssReader = reader
}

func getSPSSReader(in io.Reader) SPSSReader {
    return spssReader(in)
}

// UnmarshalFile parses the CSV from the file in the interface.
func UnmarshalFile(in *os.File, out interface{}) error {
    return Unmarshal(in, out)
}

// UnmarshalString parses the CSV from the string in the interface.
func UnmarshalString(in string, out interface{}) error {
    return Unmarshal(strings.NewReader(in), out)
}

// UnmarshalBytes parses the CSV from the bytes in the interface.
func UnmarshalBytes(in []byte, out interface{}) error {
    return Unmarshal(bytes.NewReader(in), out)
}

// Unmarshal parses the CSV from the reader in the interface.
func Unmarshal(in io.Reader, out interface{}) error {
    return readTo(newSimpleDecoderFromReader(in), out)
}
