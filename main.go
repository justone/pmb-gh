package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Println(string(body))
		go func() {
			pmb.SendNotificationWithLevel(conn, "Test message", 3)
		}()
	})
	http.ListenAndServe(":3000", nil)
}
