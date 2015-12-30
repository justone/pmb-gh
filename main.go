package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/justone/pmb/api"
)

func main() {

	bus := pmb.GetPMB("")
	id := pmb.GenerateRandomID("github")

	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	conn, err := bus.ConnectClient(id, false)
	if err != nil {
		logrus.Warnf("Error connecting to PMB: %s", err)
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read request body: "+err.Error(), http.StatusInternalServerError)
			return
		}
		var eventName string
		if eventHeaders, ok := r.Header["X-Github-Event"]; ok {
			eventName = eventHeaders[0]
		} else {
			logrus.Warnf("Github event name not found")
			return
		}

		eventJSON := string(body)

		notification, err := parseEvent(eventName, eventJSON)
		if err != nil {
			logrus.Warnf("Unable to parse event %s: %s, body: %s", eventName, err, eventJSON)
			return
		}

		go func() {
			pmb.SendNotification(conn, *notification)
		}()
	})
	http.ListenAndServe("0.0.0.0:3000", nil)
}
