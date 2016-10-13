package main

import (
	"fmt"
	"strings"

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

	skip := false

	switch name {
	case "create":
		ref_type, err := tree.Get("ref_type").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get ref_type: %s", err)
		}
		ref, err := tree.Get("ref").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get ref: %s", err)
		}
		message = fmt.Sprintf("New %s (%s) for %s by %s.", ref_type, ref, repo, login)
		base_url, err := tree.Get("repository").Get("html_url").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
		url = fmt.Sprintf("%s/tree/%s", base_url, ref)
	case "delete":
		ref_type, err := tree.Get("ref_type").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get ref_type: %s", err)
		}
		ref, err := tree.Get("ref").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get ref: %s", err)
		}
		if strings.HasPrefix(ref, "refs/heads/") {
			ref = strings.TrimPrefix(ref, "refs/heads/")
		}
		message = fmt.Sprintf("Delete %s (%s) for %s by %s.", ref_type, ref, repo, login)
		base_url, err := tree.Get("repository").Get("html_url").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
		url = fmt.Sprintf("%s/tree/%s", base_url, ref)
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
	case "push":
		commits, err := tree.Get("commits").Len()
		if err != nil {
			return nil, fmt.Errorf("Unable to get commits: %s", err)
		}
		// skip notification this is new, the create event will suffice
		if commits == 0 {
			skip = true
		}
		ref, err := tree.Get("ref").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get ref: %s", err)
		}
		if strings.HasPrefix(ref, "refs/heads/") {
			ref = strings.TrimPrefix(ref, "refs/heads/")
		}
		message = fmt.Sprintf("Push %d commit(s) to %s in %s by %s.", commits, ref, repo, login)
		url, err = tree.Get("compare").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
	case "pull_request":
		action, err := tree.Get("action").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get action: %s", err)
		}
		issue, err := tree.Get("pull_request").Get("number").Number()
		if err != nil {
			return nil, fmt.Errorf("Unable to get pull_request number: %s", err)
		}
		title, err := tree.Get("pull_request").Get("title").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get pull_request title: %s", err)
		}
		body, err := tree.Get("pull_request").Get("body").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get body: %s", err)
		}
		message = fmt.Sprintf(
			"Pull request %s #%d (%s) on %s by %s: %s",
			action,
			int(issue),
			truncate(title, 20),
			repo,
			login,
			truncate(body, 40))
		url, err = tree.Get("pull_request").Get("html_url").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
	case "issue_comment":
		action, err := tree.Get("action").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get action: %s", err)
		}
		issue, err := tree.Get("issue").Get("number").Number()
		if err != nil {
			return nil, fmt.Errorf("Unable to get issue number: %s", err)
		}
		issue_title, err := tree.Get("issue").Get("title").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get issue title: %s", err)
		}
		body, err := tree.Get("comment").Get("body").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get body: %s", err)
		}
		message = fmt.Sprintf(
			"Comment %s on issue %d (%s) on %s by %s: %s",
			action,
			int(issue),
			truncate(issue_title, 20),
			repo,
			login,
			truncate(body, 40))
		url, err = tree.Get("comment").Get("html_url").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
	case "pull_request_review":
		action, err := tree.Get("action").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get action: %s", err)
		}
		state, err := tree.Get("review").Get("state").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get state: %s", err)
		}
		issue, err := tree.Get("pull_request").Get("number").Number()
		if err != nil {
			return nil, fmt.Errorf("Unable to get pull_request number: %s", err)
		}
		title, err := tree.Get("pull_request").Get("title").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get pull_request title: %s", err)
		}
		body, err := tree.Get("review").Get("body").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get body: %s", err)
		}
		message = fmt.Sprintf(
			"PR Review %s (%s) on issue %d (%s) on %s by %s: %s",
			action,
			state,
			int(issue),
			truncate(title, 20),
			repo,
			login,
			truncate(body, 40))
		url, err = tree.Get("review").Get("html_url").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
	case "pull_request_review_comment":
		action, err := tree.Get("action").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get action: %s", err)
		}
		issue, err := tree.Get("pull_request").Get("number").Number()
		if err != nil {
			return nil, fmt.Errorf("Unable to get pull_request number: %s", err)
		}
		title, err := tree.Get("pull_request").Get("title").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get pull_request title: %s", err)
		}
		body, err := tree.Get("comment").Get("body").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get body: %s", err)
		}
		message = fmt.Sprintf(
			"PR Comment %s on issue %d (%s) on %s by %s: %s",
			action,
			int(issue),
			truncate(title, 20),
			repo,
			login,
			truncate(body, 40))
		url, err = tree.Get("comment").Get("html_url").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
	case "issues":
		action, err := tree.Get("action").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get action: %s", err)
		}
		issue, err := tree.Get("issue").Get("number").Number()
		if err != nil {
			return nil, fmt.Errorf("Unable to get issue number: %s", err)
		}
		issue_title, err := tree.Get("issue").Get("title").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get issue title: %s", err)
		}
		message = fmt.Sprintf(
			"Issue %d (%s) %s on %s by %s.",
			int(issue),
			truncate(issue_title, 20),
			action,
			repo,
			login)
		url, err = tree.Get("issue").Get("html_url").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
	case "ping":
		zen, err := tree.Get("zen").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get zen: %s", err)
		}

		message = fmt.Sprintf("Ping for %s by %s. Zen: %s", repo, login, zen)
		url, err = tree.Get("repository").Get("html_url").String()
		if err != nil {
			return nil, fmt.Errorf("Unable to get url: %s", err)
		}
	default:
		message = fmt.Sprintf("Unhandled event %s for %s by %s.", name, repo, login)
		url = ""
	}

	logrus.Debugf("message: %s", message)
	logrus.Debugf("url: %s", url)

	if skip {
		return nil, nil
	}
	return &pmb.Notification{Message: message, URL: url}, nil
}

func truncate(data string, length int) string {
	if len(data) > length {
		return fmt.Sprintf("%s...", data[0:length])
	}
	return data
}
