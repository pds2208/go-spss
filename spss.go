package spss

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
var spssWriter = DefaultSPSSWriter

func DefaultSPSSReader(in string) SPSSReader {
	return NewReader(in)
}

func SetSPSSReader(reader func(string) SPSSReader) {
	spssReader = reader
}

func getSPSSReader(in string) SPSSReader {
	return spssReader(in)
}

func ReadFromSPSS(in string, out interface{}) error {
	r := Import(in)
	if r != 0 {
		panic("Parse of " + in + " failed")
	}
	return readTo(newSimpleDecoderFromReader(in), out)
}

func SetSPSSWriter(writer func(interface{}) SPSSWriter) {
	spssWriter = writer
}

func DefaultSPSSWriter(in interface{}) SPSSWriter {
	return FileOutput{in.(string)}
}

func WriteToSPSS(out string, in interface{}) error {
	return spssWriter(out).Write()(in)
}
