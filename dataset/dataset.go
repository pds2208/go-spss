package dataset

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"github.com/deepilla/sqlitemeta"
	_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
	spss "go-spss"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/sqlite"
)

var globalLock = sync.Mutex{}

func init() {
}

type Dataset struct {
	tableName string
	tableMeta map[string]reflect.Kind
	DB        sqlbuilder.Database
	conn      *sql.DB
	mux       sync.Mutex
}

var settings = sqlite.ConnectionURL{
	Database: "LFS.db",
	Options: map[string]string{
		"cache":        "shared",
		"_synchronous": "OFF", // when not using memory: we don't need this
		"_journal":     "WAL", // much, MUCH faster
		//"mode":  "memory", // memory=prod otherwise debug so we can see the file
	},
}

func NewDataset(name string) (*Dataset, error) {
	globalLock.Lock()
	defer globalLock.Unlock()

	sess, err := sqlite.Open(settings)

	if err != nil {
		panic(err)
	}

	conn := sess.Driver().(*sql.DB)

	_, _ = sess.Exec(fmt.Sprintf("drop table if exists %s", name))

	_, err = sess.Exec(fmt.Sprintf("create table %s (Row INTEGER PRIMARY KEY)", name))
	if err != nil {
		return nil, fmt.Errorf(" -> NewDataset: cannot create table: %s, error: %s", name, err)
	}

	mux := sync.Mutex{}
	return &Dataset{name, nil, sess, conn, mux}, nil
}

func (d Dataset) Close() {
	_ = d.DB.Close()
}

func (d Dataset) AddColumn(name string, columnType spss.ColumnTypes) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	sqlStmt := fmt.Sprintf("alter table %s add %s %s", d.tableName, name, columnType)
	_, err := d.DB.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf(" -> AddColumn: cannot insert column: %s", err)
	}
	return nil
}

func (d Dataset) Insert(values interface{}) (err error) {
	q := d.DB.InsertInto(d.tableName).Values(values)
	_, err = q.Exec()
	if err != nil {
		return fmt.Errorf(" -> Insert: cannot insert row: %s", err)
	}
	return
}

func (d Dataset) Head(max ...int) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	var maxItems = 5
	if max != nil {
		maxItems = max[0]
	}

	table := tablewriter.NewWriter(os.Stdout)

	var sqlStmt = fmt.Sprintf("select * from %s limit %d", d.tableName, maxItems)
	rows, err := d.DB.Query(sqlStmt)
	if err != nil {
		return fmt.Errorf(" -> Head: Query() failed: %s", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf(" -> Head: select failed on columns: %s", err)
	}

	vals := make([]interface{}, len(cols))
	var header []string
	for i, n := range cols {
		vals[i] = new(sql.RawBytes)
		header = append(header, n)
	}
	table.SetHeader(header)

	for rows.Next() {
		err = rows.Scan(vals...)

		var rowItems []string
		for col := 0; col < len(vals); col++ {
			res := vals[col]
			b := res.(*sql.RawBytes)
			rowItems = append(rowItems, string(*b))
		}
		table.Append(rowItems)
	}

	j := fmt.Sprintf("%d Rows(s)\n", table.NumLines())
	table.SetCaption(true, j)
	table.Render()
	return nil
}

func (d Dataset) NumColumns() int {
	return len(d.columnMetadata())
}

func (d Dataset) NumRows() (count int) {
	row, _ := d.DB.QueryRow(fmt.Sprintf("select count(*) from %s", d.tableName))
	_ = row.Scan(&count)
	return
}

// helper functions

type orderedColumns = map[int]sqlitemeta.Column

// ensure table is created with existing column order
func (d Dataset) orderedColumns() (ordered orderedColumns) {
	ordered = map[int]sqlitemeta.Column{}

	res, err := sqlitemeta.Columns(d.conn, d.tableName)
	if err != nil {
		panic(fmt.Sprintf(" -> orderedColumns: cannot get metadata: %s", err))
	}
	for _, j := range res {
		ordered[j.ID] = j
	}
	return
}

type columnInfo map[string]string

func (d Dataset) columnMetadata() (colLookup columnInfo) {

	res, err := sqlitemeta.Columns(d.conn, d.tableName)
	if err != nil {
		panic(fmt.Sprintf(" -> columnMetadata: cannot get metadata for: %s", err))
	}

	colLookup = map[string]string{}

	for _, col := range res {
		colLookup[col.Name] = col.Type
	}
	return colLookup
}

func (d Dataset) doesColumnExist(name string) (bool, columnInfo) {
	var colLookup = d.columnMetadata()
	if _, ok := colLookup[name]; !ok {
		return false, nil
	}
	return true, colLookup
}

func (d Dataset) Mean(col string) (res float64, err error) {
	ok, colLookup := d.doesColumnExist(col)
	if !ok {
		return 0.0, errors.New(fmt.Sprintf(" -> Mean: column %s does not exist", col))
	}

	if colLookup[col] == string(spss.STRING) {
		return 0.0, errors.New(fmt.Sprintf(" -> Mean: column %s is not numeric", col))
	}

	row, err := d.DB.QueryRow(fmt.Sprintf("select avg(%s) from %s", col, d.tableName))
	if err != nil {
		return 0.0, err
	}
	err = row.Scan(&res)
	if err != nil {
		return 0.0, err
	}
	return
}

func (d Dataset) DropColumn(column string) (err error) {
	/*
		As sql lite can't Delete columns, we work around this by doing the following:

		1. start a transaction
		2. create a temporary table as existing table minus the column we are dropping
		3. insert all rows from table into temporary table minus the column we are dropping
		4. Delete existing table
		5. re-create table
		6. insert data from temporary into table
		7. Delete temporary table
		8. commit transaction

	*/

	d.mux.Lock()
	defer d.mux.Unlock()

	ok, colLookup := d.doesColumnExist(column)
	if !ok {
		return fmt.Errorf(" -> DropColumn: column %s does not exist: %s", column, err)
	}

	// get and save existing column order
	orderedColumns := d.orderedColumns()

	var buffer bytes.Buffer
	var keys []string
	for i := 0; i < len(orderedColumns); i++ {
		if orderedColumns[i].Name != column && orderedColumns[i].Name != "Row" {
			keys = append(keys, orderedColumns[i].Name)
		}
	}

	// start transaction

	tx, err := d.DB.NewTx(nil)
	if err != nil {
		return fmt.Errorf(" -> DropColumn: cannot create transaction: %s", err)
	}

	// create temp table
	buffer.WriteString("create table t1 (")
	for i := 0; i < len(keys); i++ {
		j := fmt.Sprintf(" %s %s", keys[i], colLookup[keys[i]])
		buffer.WriteString(j)
		if i != len(keys)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(")")

	q := buffer.String()
	_, err = d.DB.Exec(q)
	if err != nil {
		return fmt.Errorf(" -> DropColumn: Exec() failed: %s", err)
	}

	// insert into temporary table
	buffer.Reset()
	buffer.WriteString("insert into t1 (")
	for i := 0; i < len(keys); i++ {
		j := fmt.Sprintf("%s", keys[i])
		buffer.WriteString(j)
		if i != len(keys)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(") select ")
	for i := 0; i < len(keys); i++ {
		j := fmt.Sprintf("%s", keys[i])
		buffer.WriteString(j)
		if i != len(keys)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(" from ")
	buffer.WriteString(fmt.Sprintf("%s", d.tableName))
	q = buffer.String()
	_, err = d.DB.Exec(q)
	if err != nil {
		return fmt.Errorf(" -> DropColumn: Exec() failed: %s", err)
	}

	// Delete existing table
	_, err = d.DB.Exec(fmt.Sprintf("drop table %s", d.tableName))
	if err != nil {
		return fmt.Errorf(" -> DropColumn: Exec() failed: %s", err)
	}

	// re-create table
	buffer.Reset()
	buffer.WriteString(fmt.Sprintf("create table %s (Row INTEGER PRIMARY KEY, ", d.tableName))

	for i := 0; i < len(keys); i++ {
		j := fmt.Sprintf(" %s %s", keys[i], colLookup[keys[i]])
		buffer.WriteString(j)
		if i != len(keys)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(")")

	q = buffer.String()
	_, err = d.DB.Exec(q)
	if err != nil {
		return fmt.Errorf(" -> DropColumn: Exec() failed: %s", err)
	}

	// insert back into the table
	buffer.Reset()
	buffer.WriteString(fmt.Sprintf("insert into %s (", d.tableName))
	for i := 0; i < len(keys); i++ {
		j := fmt.Sprintf("%s", keys[i])
		buffer.WriteString(j)
		if i != len(keys)-1 {
			buffer.WriteString(", ")
		}
	}

	buffer.WriteString(") select ")

	for i := 0; i < len(keys); i++ {
		j := fmt.Sprintf("%s", keys[i])
		buffer.WriteString(j)
		if i != len(keys)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(" from t1 ")

	q = buffer.String()
	_, err = d.DB.Exec(q)
	if err != nil {
		return fmt.Errorf(" -> DropColumn: Exec() failed: %s", err)
	}

	// Delete temporary table
	_, err = d.DB.Exec("drop table t1")
	if err != nil {
		return fmt.Errorf(" -> DropColumn: Exec() failed: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(" -> DropColumn: Commit() failed: %s", err)
	}

	return
}

func (d Dataset) DeleteWhere(where ...interface{}) (err error) {
	err = nil
	q := d.DB.DeleteFrom(d.tableName).Where(where)
	_, err = q.Exec()
	if err != nil {
		return fmt.Errorf(" -> DeleteWhere: Exec failed: %s", err)
	}
	return
}

func (d Dataset) ToCSV(fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf(" -> ToCSV: cannot open output csv file: %s", err)
	}

	defer f.Close()

	orderedColumns := d.orderedColumns()

	var buffer bytes.Buffer
	var keys []string
	for i := 0; i < len(orderedColumns); i++ {
		if orderedColumns[i].Name != "Row" {
			keys = append(keys, orderedColumns[i].Name)
		}
	}

	for i := 0; i < len(keys); i++ {
		j := fmt.Sprintf("%s", keys[i])
		buffer.WriteString(j)
		if i != len(keys)-1 {
			buffer.WriteString(",")
		} else {
			buffer.WriteString("\n")
		}
	}

	q := buffer.String()

	_, err = f.WriteString(q)
	if err != nil {
		return fmt.Errorf(" -> ToCSV: write to file: %s failed: %s", fileName, err)
	}

	col := d.DB.Collection(d.tableName)
	res := col.Find()
	defer res.Close()
	var dat map[string]interface{}

	for res.Next(&dat) {
		buffer.Reset()

		orderedColumns := d.orderedColumns()
		var keys []string
		for i := 0; i < len(orderedColumns); i++ {
			if orderedColumns[i].Name != "Row" {
				keys = append(keys, orderedColumns[i].Name)
			}
		}

		for i := 0; i < len(keys); i++ {
			kind := d.tableMeta[keys[i]]
			value := dat[keys[i]]

			switch kind {
			case reflect.String:
				buffer.WriteString(fmt.Sprintf("%s", value))
			case reflect.Int8, reflect.Uint8:
				buffer.WriteString(fmt.Sprintf("%d", value))
			case reflect.Int, reflect.Int32, reflect.Uint32:
				buffer.WriteString(fmt.Sprintf("%d", value))
			case reflect.Int64, reflect.Uint64:
				buffer.WriteString(fmt.Sprintf("%d", value))
			case reflect.Float32:
				buffer.WriteString(fmt.Sprintf("%f", value))
			case reflect.Float64:
				buffer.WriteString(fmt.Sprintf("%g", value))
			default:
				return fmt.Errorf(" -> ToCSV: unknown type - possible corruption")
			}
			if i != len(keys)-1 {
				buffer.WriteString(",")
			} else {
				buffer.WriteString("\n")
			}
		}

		q := buffer.String()

		_, err = f.WriteString(q)
		if err != nil {
			return fmt.Errorf(" -> ToCSV: write to file: %s failed: %s", fileName, err)
		}
	}

	return nil
}

func FromSav(in string, out interface{}) (dataset Dataset, err error) {

	var empty Dataset

	// check out is a struct
	if reflect.ValueOf(out).Kind() != reflect.Struct {
		return empty, fmt.Errorf(" -> FromSav: %T is not a struct type", out)
	}

	spssRows, err := spss.Import(in)
	if err != nil {
		return empty, err
	}
	if len(spssRows) == 0 {
		return empty, fmt.Errorf(" -> FromSav: spss file: %s is empty", in)
	}

	_, file := filepath.Split(in)
	var extension = filepath.Ext(file)
	var name = file[0 : len(file)-len(extension)]

	d, er := NewDataset(name)
	if er != nil {
		return empty, fmt.Errorf(" -> FromSav: cannot create a new DataSet: %s", er)
	}

	tx, err := d.DB.NewTx(nil)
	if err != nil {
		return empty, fmt.Errorf(" -> FromSav: cannot create a transaction: %s", err)
	}

	t1 := reflect.TypeOf(out)

	d.tableMeta = make(map[string]reflect.Kind)

	for i := 0; i < t1.NumField(); i++ {
		a := t1.Field(i)
		d.tableMeta[a.Name] = a.Type.Kind()

		var spssType spss.ColumnTypes

		switch a.Type.Kind() {
		case reflect.String:
			spssType = spss.STRING
		case reflect.Int8, reflect.Uint8:
			spssType = spss.INT
		case reflect.Int, reflect.Int32, reflect.Uint32:
			spssType = spss.INT
		case reflect.Int64, reflect.Uint64:
			spssType = spss.INT
		case reflect.Float32:
			spssType = spss.FLOAT
		case reflect.Float64:
			spssType = spss.DOUBLE
		default:
			return empty, fmt.Errorf(" -> FromSav: cannot convert struct variable type from SPSS type")
		}

		err = d.AddColumn(a.Name, spssType)
		if err != nil {
			return empty, fmt.Errorf(" -> FromSav: cannot create column %s, of type %s", name, spssType)
		}

	}

	headers := spssRows[0]
	body := spssRows[1:]
	for _, spssRow := range body {
		row := make(map[string]interface{})

		for j := 0; j < len(spssRow)-1; j++ {
			if len(spssRow) != len(headers) {
				return empty, fmt.Errorf(" -> FromSav: header is out of alignment with row. row size: %d, column size: %d\n", len(spssRow), len(headers))
			}
			header := headers[j]
			// extract the columns we are interested in
			if _, ok := d.tableMeta[headers[j]]; !ok {
				continue
			}
			row[header] = spssRow[j]
		}

		err = d.Insert(row)
		if err != nil {
			return empty, fmt.Errorf(" -> FromSav: cannot create row: %s", err)
		}

	}

	err = tx.Commit()
	if err != nil {
		return empty, fmt.Errorf(" -> FromSav: commit transaction failed: %s", err)
	}

	return *d, nil
}
