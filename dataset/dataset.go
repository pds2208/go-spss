package dataset

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	spss "go-spss"
	"os"
	"strconv"
	"sync"
)

func init() {

}

type ColumnInfo struct {
	position   int
	name       string
	columnType spss.ColumnType
}

type DataItem = map[string][]interface{} // column name / data
type Column = map[string]ColumnInfo
type RowData struct {
	name  string
	value interface{}
}
type Row struct {
	row map[int][]RowData // row_number, values
}

type Dataset struct {
	data        DataItem
	rowCount    int
	columnCount int
	columnInfo  Column
	mux         sync.Mutex
}

func NewDataset() *Dataset {
	data := make(map[string][]interface{})
	col := make(map[string]ColumnInfo)
	mux := sync.Mutex{}
	return &Dataset{data, 0, 0, col, mux}
}

func (d Dataset) rows() Row {
	// we store a column view of data so need to convert to row view
	// by iterating over the columns
	var rows Row
	rows.row = make(map[int][]RowData, 0)
	var columnNames = make([]string, d.columnCount)

	for _, v := range d.columnInfo { // ensure columns are in correct position
		columnNames[v.position] = v.name
	}

	for rowNumber := 0; rowNumber <= d.rowCount-1; rowNumber++ {
		rows.row[rowNumber] = make([]RowData, 0)
		for _, colName := range columnNames {
			r := RowData{colName, d.data[colName][rowNumber]}
			rows.row[rowNumber] = append(rows.row[rowNumber], r)
		}
	}

	return rows
}

func (d *Dataset) AddRow(row map[string]interface{}) error {
	if len(d.columnInfo) == 0 {
		return fmt.Errorf("dataset has no columns. add columns first")
	}

	if len(row) != len(d.columnInfo) {
		return fmt.Errorf("%d columns required for row", len(d.columnInfo))
	}

	d.mux.Lock()

	for col, value := range row {
		switch typ := d.columnInfo[col].columnType.As(); typ {
		case spss.ReadstatTypeInt8, spss.ReadstatTypeInt16, spss.ReadstatTypeInt32:
			d.data[col] = append(d.data[col], value.(int))
		case spss.ReadstatTypeFloat:
			d.data[col] = append(d.data[col], value.(float32))
		case spss.ReadstatTypeDouble:
			d.data[col] = append(d.data[col], value.(float64))
		case spss.ReadstatTypeString:
			d.data[col] = append(d.data[col], value.(string))
		}
	}

	d.rowCount++
	d.mux.Unlock()
	return nil
}

func (d Dataset) Head(max ...int) {

	d.mux.Lock()
	var maxItems = 5
	if max != nil {
		maxItems = max[0]
	}

	table := tablewriter.NewWriter(os.Stdout)
	var header = make([]string, d.columnCount)
	for _, v := range d.columnInfo {
		header[v.position] = v.name
	}
	table.SetHeader(header)

	if maxItems > d.rowCount {
		maxItems = d.rowCount
	}

	for i := 0; i < len(d.rows().row); i++ {
		if i >= maxItems {
			break
		}

		var rowItems = make([]string, d.columnCount)
		for col := 0; col <= d.columnCount-1; col++ {

			row := d.rows().row[i][col]

			switch typ := d.columnInfo[row.name].columnType.As(); typ {
			case spss.ReadstatTypeInt8, spss.ReadstatTypeInt16, spss.ReadstatTypeInt32:
				rowItems[col] = strconv.Itoa(row.value.(int))
			case spss.ReadstatTypeFloat:
				rowItems[col] = strconv.FormatFloat(row.value.(float64), 'f', -1, 32)
			case spss.ReadstatTypeDouble:
				rowItems[col] = strconv.FormatFloat(row.value.(float64), 'f', -1, 64)
			case spss.ReadstatTypeString:
				rowItems[col] = row.value.(string)
			}
		}

		table.Append(rowItems)

	}

	j := fmt.Sprintf("%d Row(s)\n", table.NumLines())
	table.SetCaption(true, j)
	table.Render()
	d.mux.Unlock()
}

func (d *Dataset) AddColumn(name string, columnType spss.ColumnType) {
	d.mux.Lock()

	c := ColumnInfo{
		position:   d.columnCount,
		name:       name,
		columnType: columnType,
	}

	d.columnCount++
	d.columnInfo[name] = c

	d.mux.Unlock()
}

func (d *Dataset) DeleteColumn(name string) {
	d.mux.Lock()

	d.columnCount--
	delete(d.data, name)
	delete(d.columnInfo, name)

	var columnNames = make([]string, d.columnCount)
	var i = 0
	for _, v := range d.columnInfo { // ensure columns are in correct position
		columnNames[i] = v.name
		i++
	}

	// a range on a map can return data in any order so we sort it into an array by position first
	for i := 0; i < len(columnNames); i++ {
		name := columnNames[i]
		col := d.columnInfo[name]
		d.columnInfo[name] = ColumnInfo{i, name, col.columnType}
	}

	d.mux.Unlock()
}

func shiftItems() {

}

func (d Dataset) Column(col string) ([]interface{}, error) {
	d.mux.Lock()
	if val, ok := d.data[col]; ok {
		d.mux.Unlock()
		return val, nil
	}
	d.mux.Unlock()
	return nil, fmt.Errorf("requested column: %s not found", col)
}

func (d Dataset) numColumns() int {
	return len(d.columnInfo)
}

func (d Dataset) numRows() int {
	return len(d.data)
}

func (d Dataset) Mean(col string) (res float64, err error) {

	if _, ok := d.columnInfo[col]; !ok {
		return 0.0, fmt.Errorf("requested column: %s not found", col)
	}

	if !d.columnInfo[col].columnType.IsNumeric() {
		return 0.0, fmt.Errorf("requested column: %s is not numeric", col)
	}

	res = 0.0
	b := d.data[col]
	for _, item := range b {
		res += item.(float64)
	}

	return res / float64(len(b)), nil
}

func (d Dataset) ReadFromSAV(file string) error {
	return nil
}
