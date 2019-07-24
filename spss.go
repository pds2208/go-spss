package spss

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
