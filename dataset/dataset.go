package dataset

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/deepilla/sqlitemeta"
	_ "github.com/mattn/go-sqlite3"
	"github.com/olekukonko/tablewriter"
	spss "go-spss"
	"log"
	"os"
	"sync"
)

var globalLock = sync.Mutex{}

func init() {
	//db, err := sql.Open("sqlite3", ":memory")
	//db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	_, err := sql.Open("sqlite3", "LFS.db")
	if err != nil {
		panic(err)
	}
}

type Dataset struct {
	dbName string
	db     *sql.DB
	mux    sync.Mutex
}

func NewDataset(name string) (*Dataset, error) {
	globalLock.Lock()
	defer globalLock.Unlock()

	//db, err := sql.Open("sqlite3", ":memory")
	//db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	db, err := sql.Open("sqlite3", "LFS.db")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	sqlStmt := fmt.Sprintf("drop table if exists %s", name)
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}

	sqlStmt = fmt.Sprintf("create table %s (Row INTEGER PRIMARY KEY)", name)
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return nil, err
	}

	mux := sync.Mutex{}
	return &Dataset{name, db, mux}, nil
}

func (d Dataset) Close() {
	_ = d.db.Close()
}

type Row map[string]interface{}

func (d Dataset) Insert() *insert {
	return &insert{d}
}

func (d Dataset) Drop() *drop {
	return &drop{d}
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
