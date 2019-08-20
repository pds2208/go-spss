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
	"reflect"
	"sync"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/sqlite"
)

var globalLock = sync.Mutex{}

func init() {
}

type Dataset struct {
	dbName string
	DB     sqlbuilder.Database
	conn   *sql.DB
	mux    sync.Mutex
}

//db, err := sql.Open("sqlite3", ":memory")
//db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
var settings = sqlite.ConnectionURL{
	Database: "LFS.db", // file::memory:?cache=shared
}

func NewDataset(name string) (*Dataset, error) {
	globalLock.Lock()
	defer globalLock.Unlock()

	sess, err := sqlite.Open(settings)

	if err != nil {
		panic(err)
	}

	conn := sess.Driver().(*sql.DB)

	_, err = sess.Exec(fmt.Sprintf("drop table if exists %s", name))
	if err != nil {
		panic(err)
	}
	_, err = sess.Exec(fmt.Sprintf("create table %s (Row INTEGER PRIMARY KEY)", name))
	if err != nil {
		panic(err)
	}

	mux := sync.Mutex{}
	return &Dataset{name, sess, conn, mux}, nil
}

func (d Dataset) Close() {
	_ = d.DB.Close()
}

func (d Dataset) AddColumn(name string, columnType spss.ColumnTypes) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	sqlStmt := fmt.Sprintf("alter table %s add %s %s", d.dbName, name, columnType)
	_, err := d.DB.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return errors.New("invalid datatype in AddColumn")
	}
	return nil
}

func (d Dataset) Insert(values interface{}) (err error) {
	q := d.DB.InsertInto(d.dbName).Values(values)
	_, err = q.Exec()
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

	var sqlStmt = fmt.Sprintf("select * from %s limit %d", d.dbName, maxItems)
	rows, err := d.DB.Query(sqlStmt)
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

func (d Dataset) NumColumns() int {
	return len(d.columnMetadata())
}

func (d Dataset) NumRows() (count int) {
	row, _ := d.DB.QueryRow(fmt.Sprintf("select count(rowid) from %s", d.dbName))
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

	res, err := sqlitemeta.Columns(d.conn, d.dbName)
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

	res, err := sqlitemeta.Columns(d.conn, d.dbName)
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

	row, err := d.DB.QueryRow(fmt.Sprintf("select avg(%s) from %s", col, d.dbName))
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
		j := fmt.Sprintf("Delete column: column %s does not exist", column)
		return errors.New(j)
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
	_, err = d.DB.Exec(q)
	if err != nil {
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
	_, err = d.DB.Exec(q)
	if err != nil {
		return
	}

	// Delete existing table
	_, err = d.DB.Exec(fmt.Sprintf("drop table %s", d.dbName))
	if err != nil {
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
	_, err = d.DB.Exec(q)
	if err != nil {
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
	_, err = d.DB.Exec(q)
	if err != nil {
		return
	}

	// Delete temporary table
	_, err = d.DB.Exec("drop table t1")
	if err != nil {
		return
	}

	err = tx.Commit()
	return
}

func (d Dataset) DeleteWhere(where ...interface{}) (err error) {
	err = nil
	q := d.DB.DeleteFrom(d.dbName).Where(where)
	_, err = q.Exec()
	return
}

func FromSav(in string, out interface{}) (dataset Dataset, err error) {

	var empty Dataset

	if _, err := os.Stat(in); os.IsNotExist(err) {
		return empty, fmt.Errorf(" -> FromSav: file %s not found", in)
	}

	// check out is a struct
	if reflect.ValueOf(out).Kind() != reflect.Struct {
		return empty, fmt.Errorf(" -> FromSav: %T is not a struct type", out)
	}

	//err = spss.ReadFromSPSSFile(in, out)
	//if err != nil {
	//    return empty, err
	//}

	spssRows, err := spss.Import(in)
	if err != nil {
		return empty, err
	}
	if len(spssRows) == 0 {
		return empty, fmt.Errorf("spss file: %s is empty", in)
	}

	d, er := NewDataset(in)
	if er != nil {
		return empty, errors.New(" -> FromSav: cannot create a new DataSet")
	}

	d.mux.Lock()
	defer d.mux.Unlock()

	t1 := reflect.TypeOf(out)

	for i := 0; i < t1.NumField(); i++ {
		a := t1.Field(i)
		name := t1.Name()

		var spssType spss.ColumnTypes

		switch a.Type.Kind() {
		case reflect.String:
			spssType = spss.STRING
		case reflect.Int8, reflect.Uint8:
			spssType = spss.INT
		case reflect.Int, reflect.Int32, reflect.Uint32:
			spssType = spss.INT
		case reflect.Float32:
			spssType = spss.FLOAT
		case reflect.Float64:
			spssType = spss.DOUBLE
		default:
			return empty, fmt.Errorf("cannot convert type for struct variable into SPSS type")
		}

		err = d.AddColumn(name, spssType)
		if err != nil {
			return empty, fmt.Errorf(" -> FromSav: cannot create column %s, of type %s", name, spssType)
		}

	}

	headers := spssRows[0]
	body := spssRows[1:]

	for _, spssRow := range body {
		row := make(map[string]interface{})
		for j, columnContent := range spssRow {
			row[headers[j]] = columnContent
		}
		_ = dataset.Insert(row)
	}

	return *d, nil
}
