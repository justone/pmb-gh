package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/justone/pmb/api"
)

var opts struct {
	Verbose bool     `short:"v" long:"verbose" description:"Show verbose debug information."`
	Primary string   `short:"m" long:"pmb-uri" description:"Primary PMB URI."`
	Ignore  []string `short:"i" long:"ignore" description:"Github username to ignore."`
	Level   float64  `short:"l" long:"level" description:"Level at which to send notifications." default:"4"`
	Host    string   `short:"h" long:"host" description:"Host to listen on." default:"0.0.0.0"`
	Port    string   `short:"p" long:"port" description:"Port to listen on." default:"3000"`
}

func main() {

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	ignoreUsers := make(map[string]bool)
	for _, u := range opts.Ignore {
		ignoreUsers[u] = true
	}

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
		var ip string
		if realIP, ok := r.Header["X-Real-Ip"]; ok {
			ip = realIP[0]
		} else {
			ip = r.RemoteAddr
		}
		logrus.Infof(strings.Join([]string{r.RequestURI, ip, r.Method, fmt.Sprintf("%s", r.Header)}, " "))

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

		notification, login, err := parseEvent(eventName, eventJSON)

		if _, ok := ignoreUsers[login]; ok {
			logrus.Warnf(fmt.Sprintf("ignoring notification from %s", login))
			return
		}
		if err != nil {
			logrus.Warnf("Unable to parse event %s: %s, body: %s", eventName, err, eventJSON)
			return
		}

		if notification == nil {
			logrus.Warnf("skipping notification")
			return
		}

		notification.Level = opts.Level

		logrus.Infof("Sending notification: %s", notification)
		go func() {
			pmb.SendNotification(conn, *notification)
		}()
	})

	http.ListenAndServe(fmt.Sprintf("%s:%s", opts.Host, opts.Port), nil)
}
