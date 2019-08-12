package dataset

import (
	"fmt"
)

const (
	ReadstatTypeString    = iota
	ReadstatTypeInt8      = iota
	ReadstatTypeInt16     = iota
	ReadstatTypeInt32     = iota
	ReadstatTypeFloat     = iota
	ReadstatTypeDouble    = iota
	ReadstatTypeStringRef = iota
)

type DataItem struct {
	column Column
	row    Row
}

type Dataset struct {
	columns []Column
	rows    []Row
}

type Row struct {
	rowNumber int
	rowValue  []interface{}
}

type Column struct {
	columnNumber int
	name         string
	columnType   int
}

func (d Dataset) Column(col int) ([]interface{}, error) {
	if col > len(d.columns) || col < 0 {
		return nil, fmt.Errorf("requested column: %d is out of range (total columns: %d)", col, len(d.columns))
	}

	var cols []interface{}
	for _, row := range d.rows {
		cols = append(cols, row.rowValue[col])

	}
	return cols, nil
}

func (d Dataset) Cell(row, col int) (DataItem, error) {
	if row > len(d.rows) || row < 0 {
		return DataItem{}, fmt.Errorf("requested row: %d is out of range (total rows: %d)", row, len(d.rows))
	}
	if col > len(d.columns) || col < 0 {
		return DataItem{}, fmt.Errorf("requested column: %d is out of range (total columns: %d)", col, len(d.columns))
	}
	return DataItem{d.columns[col], d.rows[row]}, nil
}

func (d Dataset) numRows() int {
	return len(d.rows)
}

func (d Dataset) Mean(column int) (res float64, err error) {

	if d.columns[1].columnType != ReadstatTypeInt8 && d.columns[1].columnType != ReadstatTypeInt16 &&
		d.columns[1].columnType != ReadstatTypeInt32 && d.columns[1].columnType != ReadstatTypeFloat &&
		d.columns[1].columnType != ReadstatTypeDouble {

		return 0.0, fmt.Errorf("requested column: %d is not numeric", column)
	}

	columns, err := d.Column(column)

	if err != nil {
		return 0.0, err
	}

	res = 0.0
	for _, j := range columns {
		res += j.(float64)
	}
	return res / float64(len(columns)), nil
}

func (d Dataset) ReadFromSAV(file string) error {
	return nil
}
