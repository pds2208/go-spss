package spss

import (
	"bytes"
	"encoding/csv"
	"io"
	"os"
)

/*
Use of this source code is governed by a MIT license
The license can be found in the LICENSE file.

The GoSPSS package aims to provide  SPSS serialisation and deserialisation
*/

var FailIfUnmatchedStructTags = true

var FailIfDoubleHeaderNames = false

var ShouldAlignDuplicateHeadersWithStructFieldOrder = false

var TagSeparator = ","

var spssReader = DefaultSPSSReader
var selfCSVWriter = DefaultSPSSWriter

func DefaultSPSSWriter(out io.Writer) *SafeSPSSWriter {
	return NewSafeSPSSWriter(csv.NewWriter(out))
}

func DefaultSPSSReader(in string) SPSSReader {
	return NewReader(in)
}

func SetSPSSReader(reader func(string) SPSSReader) {
	spssReader = reader
}

func getSPSSReader(in string) SPSSReader {
	return spssReader(in)
}

func UnmarshalFile(in string, out interface{}) error {
	r := Import(in)
	if r != 0 {
		panic("Parse of " + in + " failed")
	}
	return readTo(newSimpleDecoderFromReader(in), out)
}

func MarshalFile(in interface{}, file *os.File) (err error) {
	return Marshal(in, file)
}

func MarshalString(in interface{}) (out string, err error) {
	bufferString := bytes.NewBufferString(out)
	if err := Marshal(in, bufferString); err != nil {
		return "", err
	}
	return bufferString.String(), nil
}

func MarshalBytes(in interface{}) (out []byte, err error) {
	bufferString := bytes.NewBuffer(out)
	if err := Marshal(in, bufferString); err != nil {
		return nil, err
	}
	return bufferString.Bytes(), nil
}

func Marshal(in interface{}, out io.Writer) (err error) {
	writer := getSPSSWriter(out)
	return writeTo(writer, in, false)
}

func SetSPSSWriter(csvWriter func(io.Writer) *SafeSPSSWriter) {
	selfCSVWriter = csvWriter
}

func getSPSSWriter(out io.Writer) *SafeSPSSWriter {
	return selfCSVWriter(out)
}
