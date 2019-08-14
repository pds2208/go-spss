package dataset

import (
	spss "go-spss"
	"testing"
)

func Test_mean(t *testing.T) {
	var dataset = NewDataset()

	dataset.AddColumn("Name", spss.ReadstatTypeString)
	dataset.AddColumn("Address", spss.ReadstatTypeString)
	dataset.AddColumn("PostCode", spss.ReadstatTypeInt32)

	row1 := map[string]interface{}{
		"Name":     "Fred Bloggs",
		"Address":  "123 the valleys newport wales",
		"PostCode": 1908,
	}

	row2 := map[string]interface{}{
		"Name":     "Elinor Pain",
		"Address":  "Down the pub, as usual",
		"PostCode": 666,
	}
	row3 := map[string]interface{}{
		"Name":     "George",
		"Address":  "Down the pub, as usual",
		"PostCode": 666,
	}
	_ = dataset.AddRow(row1)
	_ = dataset.AddRow(row2)
	_ = dataset.AddRow(row3)

	//_ = dataset.Rows()
	dataset.Head()

	dataset.DeleteColumn("Address")
	dataset.Head()
}
