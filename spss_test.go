package spss

import (
    "fmt"
    "os"
    "testing"
)

type Client struct {
    Shiftno float64 `spss:"Shiftno"`
    Serial  float64 `spss:"Serial"`
    Version string  `spss:"Version"`
}

func Test_spss(t *testing.T) {
    clientsFile, err := os.OpenFile("testdata/ips1710bv2.sav", os.O_RDWR|os.O_CREATE, os.ModePerm)
    if err != nil {
        panic(err)
    }
    defer clientsFile.Close()

    clients := []*Client{}

    if err := UnmarshalFile(clientsFile, &clients); err != nil { // Load clients from file
        panic(err)
    }
    for _, client := range clients {
        fmt.Println("Hello", client.Serial)
    }

    if _, err := clientsFile.Seek(0, 0); err != nil { // Go to the start of the file
        panic(err)
    }

    clients = append(clients, &Client{Id: "12", Name: "John", Age: "21"}) // Add clients
    clients = append(clients, &Client{Id: "13", Name: "Fred"})
    clients = append(clients, &Client{Id: "14", Name: "James", Age: "32"})
    clients = append(clients, &Client{Id: "15", Name: "Danny"})
    csvContent, err := gocsv.MarshalString(&clients) // Get all clients as CSV string
    //err = gocsv.MarshalFile(&clients, clientsFile) // Use this to save the CSV back to the file
    if err != nil {
        panic(err)
    }
    fmt.Println(csvContent) // Display all clients as CSV string

}

