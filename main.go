package main

import (
	"fmt"
	"os"

	"github.com/justone/pmb/api"
)

func main() {

	bus := pmb.GetPMB("")
	id := pmb.GenerateRandomID("notify")

	conn, err := bus.ConnectClient(id, false)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pmb.SendNotificationWithLevel(conn, "Test message", 3)
}
