package rserve

import (
	"fmt"
	"github.com/senseyeio/roger"
	"testing"
)

var rClient roger.RClient = nil

func init() {
	r, err := roger.NewRClient("127.0.0.1", 6311)
	if err != nil {
		panic("Failed to connect")
	}
	rClient = r
}

/*
Star Rserve on linux with:
 PATH=$PATH:/usr/local/lib/R/site-library/Rserve/libs && R CMD Rserve
*/
func TestRserve(t *testing.T) {

	value, err := times(15.2, 152.917568)
	if err != nil {
		fmt.Println("Command failed: " + err.Error())
	} else {
		fmt.Println(value)
	}

}

func times(x, y float64) (float64, error) {
	str := fmt.Sprintf("times(%f, %f)", x, y)
	value, err := rClient.Eval(str)
	if err != nil {
		fmt.Println("Command failed: " + err.Error())
		return 0, err
	}
	return value.(float64), nil
}
