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
	_ = dataset.Insert(row1)
	_ = dataset.Insert(row2)
	_ = dataset.Insert(row3)

	return
}

func TestDeleteWhere(t *testing.T) {
	dataset, err := setupTable()
	if err != nil {
		panic(err)
	}

	defer dataset.Close()

	err = dataset.DeleteWhere("PostCode = ? and HowMany = ?", 667, 0)
	rows := dataset.NumRows()
	if rows != 3 {
		t.Errorf("DeleteWhere failed as NumRows is incorrect, got: %d, want: %d.", rows, 3)
	}

	err = dataset.DeleteWhere("PostCode", 667, "HowMany", 12.24)
	rows = dataset.NumRows()
	if rows != 2 {
		t.Errorf("DeleteWhere failed as NumRows is incorrect, got: %d, want: %d.", rows, 2)
	}

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

	err = dataset.DropColumn("Address")
	cols := dataset.NumColumns()
	if cols != 4 {
		t.Errorf("DropByColumn failed as NumColumns is incorrect, got: %d, want: %d.", cols, 4)
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
		t.Errorf("TestMean failed as mean value is incorrect, got: %f, want: %f.", mean, 11.24)
	}

}

func TestReadSav(t *testing.T) {

	type SpssFile struct {
		Shiftno      int    `spss:"Shiftno"`
		Serial       int64  `spss:"Serial"`
		Version      string `spss:"Version"`
		PortRoute2   int    `spss:"PortRoute2"`
		Baseport     string `spss:"Baseport"`
		PRouteLatDeg int    `spss:"PRouteLatDeg"`
		PRouteLonEW  string `spss:"PRouteLonEW"`
		DVLineName   string `spss:"DVLineName"`
		DVPortName   string `spss:"DVPortName"`
	}

	dataset, err := FromSav("../testdata/ips1710bv2.sav", SpssFile{})
	if err != nil {
		panic(err)
	}
	defer dataset.Close()

	t.Logf("Dataset Size: %d\n", dataset.NumRows())
	_ = dataset.Head(5)
}

func TestWriteCSV(t *testing.T) {

	type SpssFile struct {
		Shiftno      int    `spss:"Shiftno"`
		Serial       int64  `spss:"Serial"`
		Version      string `spss:"Version"`
		PortRoute2   int    `spss:"PortRoute2"`
		Baseport     string `spss:"Baseport"`
		PRouteLatDeg int    `spss:"PRouteLatDeg"`
		PRouteLonEW  string `spss:"PRouteLonEW"`
		DVLineName   string `spss:"DVLineName"`
		DVPortName   string `spss:"DVPortName"`
	}

	dataset, err := FromSav("../testdata/ips1710bv2.sav", SpssFile{})
	if err != nil {
		panic(err)
	}
	defer dataset.Close()

	err = dataset.ToCSV("out.csv")
	if err != nil {
		panic(err)
	}

	t.Logf("Dataset Size: %d\n", dataset.NumRows())
	_ = dataset.Head(5)
}
