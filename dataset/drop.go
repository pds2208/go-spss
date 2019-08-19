package dataset

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
)

// dataset.drop().ByColumn("column") - implies column
// dataset.drop().ByRowNumber(int)
// dataset.drop().Where(condition string)
// condition

type drop struct {
	dataset Dataset
}

func (d drop) ByRowNumber(rowNo int) (err error) {
	err = nil
	row := d.dataset.db.QueryRow(fmt.Sprintf("delete from %s where Row = %d", d.dataset.dbName, rowNo))
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	return
}

func (d drop) ByColumn(column string) (err error) {
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

	d.dataset.mux.Lock()
	defer d.dataset.mux.Unlock()

	ok, colLookup := d.dataset.doesColumnExist(column)
	if !ok {
		j := fmt.Sprintf("drop column: column %s does not exist", column)
		return errors.New(j)
	}

	// get and save existing column order
	orderedColumns := d.dataset.orderedColumns()

	var buffer bytes.Buffer
	var keys []string
	for i := 0; i < len(orderedColumns); i++ {
		if orderedColumns[i].Name != column && orderedColumns[i].Name != "Row" {
			keys = append(keys, orderedColumns[i].Name)
		}
	}

	// start transaction
	tx, err := d.dataset.db.Begin()
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
	row := d.dataset.db.QueryRow(q)
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
	buffer.WriteString(fmt.Sprintf("%s", d.dataset.dbName))
	q = buffer.String()
	row = d.dataset.db.QueryRow(q)
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	// drop existing table
	row = d.dataset.db.QueryRow(fmt.Sprintf("drop table %s", d.dataset.dbName))
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	// re-create table
	buffer.Reset()
	buffer.WriteString(fmt.Sprintf("create table %s (Row INTEGER PRIMARY KEY, ", d.dataset.dbName))

	for i := 0; i < len(keys); i++ {
		j := fmt.Sprintf(" %s %s", keys[i], colLookup[keys[i]])
		buffer.WriteString(j)
		if i != len(keys)-1 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(")")

	q = buffer.String()
	row = d.dataset.db.QueryRow(q)
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	// insert back into the table
	buffer.Reset()
	buffer.WriteString(fmt.Sprintf("insert into %s (", d.dataset.dbName))
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
	row = d.dataset.db.QueryRow(q)
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	// delete temporary table
	row = d.dataset.db.QueryRow("drop table t1")
	err = row.Scan()
	if err != sql.ErrNoRows {
		return
	}

	err = tx.Commit()
	return
}
