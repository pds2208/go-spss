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
	"log"
	"os"
	"strconv"
	"sync"
)

func init() {
}

type Dataset struct {
	dbName string
	db     *sql.DB
	mux    sync.Mutex
}

func NewDataset(name string) (*Dataset, error) {
	mux := sync.Mutex{}
	//db, err := sql.Open("sqlite3", ":memory")
	//db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	db, err := sql.Open("sqlite3", "/Users/paul/LFS.db")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	var sqlStmt = fmt.Sprintf("create table %s (Row INTEGER PRIMARY KEY)", name)
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}

	return &Dataset{name, db, mux}, nil
}

func (d Dataset) Close() {
	_ = d.db.Close()
}

type Row map[string]interface{}

func (d *Dataset) AddRow(row Row) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	var colLookup = d.columnMetadata()

	var colNames bytes.Buffer
	var colValues bytes.Buffer
	var keys []string
	for key := range row {
		keys = append(keys, key)
	}
	for i := 0; i < len(keys); i++ {
		col := keys[i]
		value := row[col]

		colNames.WriteString(col)

		switch colLookup[col] {
		case "TEXT":
			colValues.WriteString("'" + value.(string) + "'")
		case "INTEGER", "BIGINT":
			colValues.WriteString(strconv.Itoa(value.(int)))
		case "FLOAT":
			colValues.WriteString(strconv.FormatFloat(value.(float64), 'f', -1, 32))
		case "DOUBLE":
			colValues.WriteString(strconv.FormatFloat(value.(float64), 'f', -1, 64))
		}

		if i != len(keys)-1 {
			colNames.WriteString(", ")
			colValues.WriteString(", ")
		}
	}

	var sqlStmt = fmt.Sprintf("insert into %s (%s) values (%s)", d.dbName, colNames.String(), colValues.String())
	_, err := d.db.Exec(sqlStmt)
	if err != nil {
		log.Printf("Insert failed: %q: %s\n", err, sqlStmt)
		return err
	}

	return nil
}

// Need a where condition
func (d *Dataset) DeleteRow() {

}

func (d Dataset) Head(max ...int) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	var maxItems = 5
	if max != nil {
		maxItems = max[0]
	}

	table := tablewriter.NewWriter(os.Stdout)

	var sqlStmt = fmt.Sprintf("select * from %s limit %d", d.dbName, maxItems)
	rows, err := d.db.Query(sqlStmt)
	if err != nil {
		log.Printf("select failed: %q: %s\n", err, sqlStmt)
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	cols, err := rows.Columns()
	if err != nil {
		log.Printf("select failed on columns: %q: %s\n", err, sqlStmt)
		return err
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

func (d *Dataset) AddColumn(name string, columnType spss.ColumnTypes) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	sqlStmt := fmt.Sprintf("alter table %s add %s %s", d.dbName, name, columnType)
	_, err := d.db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return errors.New("invalid datatype in AddColumn")
	}
	return nil
}

/*
As sql lite can't drop columns, we work around this by doing the following:

1. start a transaction
2. create a temporary table as existing table minus the column we are dropping
3. insert all rows from table into temporary table minus the column we are dropping
4. drop existing table
5. re-create table
6. insert data from temporary into table
7. drop temporary table
8. commit transaction

*/
func (d *Dataset) DeleteColumn(name string) (err error) {
	d.mux.Lock()
	defer d.mux.Unlock()

	ok, colLookup := d.doesColumnExist(name)
	if !ok {
		j := fmt.Sprintf("drop column: column %s does not exist", name)
		return errors.New(j)
	}

	// get and save existing column order
	orderedColumns := d.orderedColumns()

	var buffer bytes.Buffer
	var keys []string
	for i := 0; i < len(orderedColumns); i++ {
		if orderedColumns[i].Name != name && orderedColumns[i].Name != "Row" {
			keys = append(keys, orderedColumns[i].Name)
		}
	}

	// start transaction
	tx, err := d.db.Begin()
	if err != nil {
		return
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
	row := d.db.QueryRow(q)
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
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
	buffer.WriteString(fmt.Sprintf("%s", d.dbName))
	q = buffer.String()
	row = d.db.QueryRow(q)
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	// drop existing table
	row = d.db.QueryRow(fmt.Sprintf("drop table %s", d.dbName))
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	// re-create table
	buffer.Reset()
	buffer.WriteString(fmt.Sprintf("create table %s (Row INTEGER PRIMARY KEY, ", d.dbName))

	for i := 0; i < len(keys); i++ {
		j := fmt.Sprintf(" %s %s", keys[i], colLookup[keys[i]])
		buffer.WriteString(j)
		if i != len(keys)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(")")

	q = buffer.String()
	row = d.db.QueryRow(q)
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	// insert back into the table
	buffer.Reset()
	buffer.WriteString(fmt.Sprintf("insert into %s (", d.dbName))
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
	row = d.db.QueryRow(q)
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	// delete temporary table
	row = d.db.QueryRow("drop table t1")
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	err = tx.Commit()
	return
}

func (d Dataset) NumColumns() int {
	return len(d.columnMetadata())
}

func (d Dataset) NumRows() (count int) {

	row := d.db.QueryRow(fmt.Sprintf("select count(rowid) from %s", d.dbName))
	switch err := row.Scan(&count); err {
	case sql.ErrNoRows:
		return 0
	case nil:
		return
	default:
		panic(err)
	}
}

// helper functions

type orderedColumns = map[int]sqlitemeta.Column

// ensure table is created with existing column order
func (d Dataset) orderedColumns() (ordered orderedColumns) {
	ordered = map[int]sqlitemeta.Column{}
	res, err := sqlitemeta.Columns(d.db, d.dbName)
	if err != nil {
		panic(fmt.Sprintf("cannot get metadata: %s", err))
	}
	for _, j := range res {
		ordered[j.ID] = j
	}
	return
}

type columnInfo map[string]string

func (d Dataset) columnMetadata() (colLookup columnInfo) {
	res, err := sqlitemeta.Columns(d.db, d.dbName)
	if err != nil {
		panic(fmt.Sprintf("cannot get metadata for: %s", err))
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
		return 0.0, errors.New(fmt.Sprintf("Mean: column %s does not exist", col))
	}

	if colLookup[col] == string(spss.STRING) {
		return 0.0, errors.New(fmt.Sprintf("Mean: column %s is not numeric", col))
	}

	row := d.db.QueryRow(fmt.Sprintf("select avg(%s) from %s", col, d.dbName))
	err = row.Scan(&res)
	if err != nil {
		return 0.0, err
	}
	return
}

func (d Dataset) ReadFromSAV(file string) error {
	return nil
}
