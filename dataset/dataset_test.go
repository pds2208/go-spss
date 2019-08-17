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

	_ = dataset.AddColumn("Name", spss.STRING)
	_ = dataset.AddColumn("Address", spss.STRING)
	_ = dataset.AddColumn("PostCode", spss.INT)
	_ = dataset.AddColumn("HowMany", spss.FLOAT)

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
	_ = dataset.AddRow(row1)
	_ = dataset.AddRow(row2)
	_ = dataset.AddRow(row3)

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

func TestDropColumn(t *testing.T) {
	dataset, err := setupTable()
	if err != nil {
		panic(err)
	}
	defer dataset.Close()

	err = dataset.DeleteColumn("Address")
	cols := dataset.NumColumns()
	if cols != 4 {
		t.Errorf("DropColumn failed as NumColumns is incorrect, got: %d, want: %d.", cols, 4)
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
