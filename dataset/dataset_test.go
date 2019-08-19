package dataset

import (
	spss "go-spss"
	"os"
	"testing"
)

func setupTable() (dataset *Dataset, err error) {
	_ = os.Remove("LFS.db")
	dataset, err = NewDataset("address")

	if err != nil {
		panic("Cannot create database")
	}

	_ = dataset.Insert().Column("Name", spss.STRING)
	_ = dataset.Insert().Column("Address", spss.STRING)
	_ = dataset.Insert().Column("PostCode", spss.INT)
	_ = dataset.Insert().Column("HowMany", spss.FLOAT)

	row1 := map[string]interface{}{
		"Name":     "Boss Lady",
		"Address":  "123 the Valleys Newport Wales",
		"PostCode": 1908,
		"HowMany":  10.24,
	}

	row2 := map[string]interface{}{
		"Name":     "Thorny El",
		"Address":  "Down the pub, as usual",
		"PostCode": 666,
		"HowMany":  11.24,
	}
	row3 := map[string]interface{}{
		"Name":     "George the Dragon",
		"Address":  "With El down the pub",
		"PostCode": 667,
		"HowMany":  12.24,
	}
	_ = dataset.Insert().Row(row1)
	_ = dataset.Insert().Row(row2)
	_ = dataset.Insert().Row(row3)

	return
}

func TestNumberRowsColumns(t *testing.T) {
	dataset, err := setupTable()
	if err != nil {
		panic(err)
	}
	defer dataset.Close()

	rows := dataset.NumRows()
	cols := dataset.NumColumns()
	if rows != 3 {
		t.Errorf("NumRows was incorrect, got: %d, want: %d.", rows, 3)
	}
	if cols != 5 {
		t.Errorf("NumColumns was incorrect, got: %d, want: %d.", cols, 5)
	}
}

func TestDropByColumn(t *testing.T) {
	dataset, err := setupTable()
	if err != nil {
		panic(err)
	}
	defer dataset.Close()

	err = dataset.Drop().ByColumn("Address")
	cols := dataset.NumColumns()
	if cols != 4 {
		t.Errorf("DropByColumn failed as NumColumns is incorrect, got: %d, want: %d.", cols, 4)
	}
}

func TestDropRow(t *testing.T) {
	dataset, err := setupTable()
	if err != nil {
		panic(err)
	}
	defer dataset.Close()

	err = dataset.Drop().ByRowNumber(1)
	rows := dataset.NumRows()
	if rows != 2 {
		t.Errorf("DropRow failed as NumRows is incorrect, got: %d, want: %d.", rows, 2)
	}
}

func TestMean(t *testing.T) {

	dataset, err := setupTable()
	if err != nil {
		panic(err)
	}

	mean, err := dataset.Mean("HowMany")
	if err != nil {
		panic(err)
	}

	if mean != 11.24 {
		t.Errorf("DropColumn failed as NumColumns is incorrect, got: %f, want: %f.", mean, 11.24)
	}

}
