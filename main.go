package main

import (
	"fmt"
	"os"

	"dead-pool/deadpool"
)

var clientID = ""
var clientSecret = ""

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: program <userID> <projectID>")
		return
	}

	deadpool.Init(clientID, clientSecret, "./dead-pool.log")
	deadpool.CheckPoolProject(os.Args[1], os.Args[2])
	deadpool.CloseLogs()
}
