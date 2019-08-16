package spss

type ColumnType int
type ColumnTypes string

const (
	ReadstatTypeString    ColumnType = iota
	ReadstatTypeInt8      ColumnType = iota
	ReadstatTypeInt16     ColumnType = iota
	ReadstatTypeInt32     ColumnType = iota
	ReadstatTypeFloat     ColumnType = iota
	ReadstatTypeDouble    ColumnType = iota
	ReadstatTypeStringRef ColumnType = iota
)

const (
	INT     ColumnTypes = "INTEGER"
	INTEGER ColumnTypes = "INTEGER"
	BIGINT  ColumnTypes = "BIGINT"
	STRING  ColumnTypes = "TEXT"
	FLOAT   ColumnTypes = "FLOAT"
	DOUBLE  ColumnTypes = "DOUBLE"
)

func (columnType ColumnType) As() ColumnType {
	return columnType
}

func (columnType ColumnType) AsInt() int {
	return int(columnType)
}

func (columnType ColumnType) IsNumeric() bool {
	if columnType != ReadstatTypeInt8 && columnType != ReadstatTypeInt16 &&
		columnType != ReadstatTypeInt32 && columnType != ReadstatTypeFloat &&
		columnType != ReadstatTypeDouble {
		return false
	}
	return true
}
