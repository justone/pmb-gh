package main

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/bmatsuo/go-jsontree"
	"github.com/justone/pmb/api"
)

func parseEvent(name string, json string) (*pmb.Notification, error) {

	var message string
	var url string

	tree := jsontree.New()
	err := tree.UnmarshalJSON([]byte(json))
	if err != nil {
		return nil, err
	}

	repo, err := tree.Get("repository").Get("full_name").String()
	if err != nil {
		return nil, fmt.Errorf("Unable to get full name of repository: %s", err)
	}
	login, err := tree.Get("sender").Get("login").String()
	if err != nil {
		return nil, fmt.Errorf("Unable to get full name of repository: %s", err)
	}

	switch name {
	case "watch":
		message = fmt.Sprintf("New star for %s by %s.", repo, login)
		url, err = tree.Get("sender").Get("html_url").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
	case "fork":
		message = fmt.Sprintf("New fork for %s by %s.", repo, login)
		url, err = tree.Get("forkee").Get("html_url").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
	default:
		return nil, fmt.Errorf("Invalid message type '%s'", name)
	}

	logrus.Debugf("message: %s", message)
	logrus.Debugf("url: %s", url)

	return &pmb.Notification{Message: message, URL: url}, nil
}
