package r

import (
	"fmt"
	"github.com/senseyeio/roger"
	"testing"
)

/*
Star Rserve on linux with:
 PATH=$PATH:/usr/local/lib/R/site-library/Rserve/libs && R CMD Rserve
*/
func TestRserve(t *testing.T) {
	rClient, err := roger.NewRClient("127.0.0.1", 6311)
	if err != nil {
		fmt.Println("Failed to connect")
		return
	}
	//
	//value, err := rClient.Eval("pi")
	//if err != nil {
	//	fmt.Println("Command failed: " + err.Error())
	//} else {
	//	fmt.Println(value) // 3.141592653589793
	//}
	//
	//helloWorld, _ := rClient.Eval("as.character('Hello World')")
	//fmt.Println(helloWorld) // Hello World
	//
	////arrChan := rClient.Evaluate("Sys.sleep(5); c(1,1)")
	////arrResponse := <-arrChan
	////arr, _ := arrResponse.GetResultObject()
	////fmt.Println(arr) // [1, 1]
	//
	//value, err := rClient.Eval("source('/home/paul/times.R')")
	//if err != nil {
	//	fmt.Println("Command failed: " + err.Error())
	//} else {
	//	//fmt.Println(value) // 3.141592653589793
	//}

	value, err := rClient.Eval("res <- as.integer(times(5,5))")
	if err != nil {
		fmt.Println("Command failed: " + err.Error())
	} else {
		fmt.Println(value)
	}

}

func getResultObject(command string) (interface{}, error) {
	client, _ := roger.NewRClient("localhost", 6311)
	return client.Eval(command)
}
