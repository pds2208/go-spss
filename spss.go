package spss

/*
Use of this source code is governed by a MIT license
The license can be found in the LICENSE file.

The go-spss package aims to provide SPSS serialisation and deserialisation
*/

var FailIfUnmatchedStructTags = true
var FailIfDoubleHeaderNames = false
var TagSeparator = ","

const EOL = "\n"

var spssReader = DefaultSPSSReader
var spssWriter = DefaultSPSSWriter

func DefaultSPSSReader(in interface{}) Reader {
	return FileInput{in.(string)}
}

func SetSPSSReader(reader func(interface{}) Reader) {
	spssReader = reader
}

func ReadFromSPSSFile(in string, out interface{}) error {
	return spssReader(in).Read(out)
}

func SetSPSSWriter(writer func(interface{}) Writer) {
	spssWriter = writer
}

func DefaultSPSSWriter(in interface{}) Writer {
	return FileOutput{in.(string)}
}

func WriteToSPSSFile(out string, in interface{}) error {
	return spssWriter(out).Write(in)
}
