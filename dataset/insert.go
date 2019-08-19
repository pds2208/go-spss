package dataset

import (
	"bytes"
	"errors"
	"fmt"
	spss "go-spss"
	"log"
	"strconv"
)

type insert struct {
	dataset Dataset
}

func (d insert) Column(name string, columnType spss.ColumnTypes) error {
	d.dataset.mux.Lock()
	defer d.dataset.mux.Unlock()

	sqlStmt := fmt.Sprintf("alter table %s add %s %s", d.dataset.dbName, name, columnType)
	_, err := d.dataset.db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return errors.New("invalid datatype in AddColumn")
	}
	return nil
}

func (d insert) Row(row Row) error {
	d.dataset.mux.Lock()
	defer d.dataset.mux.Unlock()

	var colLookup = d.dataset.columnMetadata()

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

	var sqlStmt = fmt.Sprintf("insert into %s (%s) values (%s)", d.dataset.dbName, colNames.String(), colValues.String())
	_, err := d.dataset.db.Exec(sqlStmt)
	if err != nil {
		log.Printf("insert failed: %q: %s\n", err, sqlStmt)
		return err
	}

	return nil
}
