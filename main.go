package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/justone/pmb/api"
)

var opts struct {
	Verbose bool    `short:"v" long:"verbose" description:"Show verbose debug information."`
	Primary string  `short:"m" long:"pmb-uri" description:"Primary PMB URI."`
	Level   float64 `short:"l" long:"level" description:"Level at which to send notifications." default:"4"`
	Host    string  `short:"h" long:"host" description:"Host to listen on." default:"0.0.0.0"`
	Port    string  `short:"p" long:"port" description:"Port to listen on." default:"3000"`
}

func main() {

	args, err := flags.Parse(&opts)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	fmt.Println(args)

	bus := pmb.GetPMB(opts.Primary)
	id := pmb.GenerateRandomID("github")

	if opts.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
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

		notification.Level = opts.Level

		go func() {
			pmb.SendNotification(conn, *notification)
		}()
	})

	http.ListenAndServe(fmt.Sprintf("%s:%s", opts.Host, opts.Port), nil)
}
