package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hugbotme/bot-go/config"
	"github.com/hugbotme/bot-go/parser"
	"github.com/hugbotme/bot-go/twitter"
	"log"
	netUrl "net/url"
	"strings"
)

type GitHubURL struct {
	URL        *netUrl.URL
	Owner      string
	Repository string
}

func processHug(url *Hug, config *config.Configuration, stopWordsFile string, probableWordsFile string) {
	fmt.Println("Parsing repository: " + url.URL)

	gitHubUrl, err := ParseGitHubURL(url.URL)

	if err != nil {
		log.Printf("Error during url parsing: %v\n", err)
		return
	}

	parser := parser.NewParser(gitHubUrl.Owner, gitHubUrl.Repository, config)

	repo, _, err := parser.ForkRepository()
	if err != nil {
		log.Printf("Error during fork: %v\n", err)
		return
	}

	log.Printf("Forked repo:" + *repo.CloneURL)
	log.Printf("Clone path:" + parser.GetClonedProjectsPath() + parser.GetRepositoryname())

	err = parser.Clone(repo)
	if err != nil {
		log.Printf("Error during clone: %v\n", err)
		return
	}

	// jvt: @todo this could all be streamed through memory as a byte stream
	filename, lines, err := parser.GetReadme()

	var buffer bytes.Buffer
	if err != nil {
		log.Printf("Error reading README: %v\n", err)
	} else {
		for _, line := range lines {
			//fmt.Println(i, line)
			buffer.WriteString(line)
		}

		processor, err := newSpellCheckFileProcessor(stopWordsFile, probableWordsFile)
		if err != nil {
			fmt.Errorf("could not get speller: %s", err.Error())
			return
		}

		content := processor.processContent([]byte(buffer.String()))

		branchname := "bugfix"

		// TODO: ERROR HANDLING
		branch, err := parser.CreateBranch(branchname)
		if err != nil {
			log.Println("CreateBranch failed:", err)
			return
		}
		err = parser.CommitFile(branch, branchname, filename, content, "Fixing some typos")
		if err != nil {
			log.Println("Commit failed:", err)
			return
		}

		tweet_add := ""
		if len(url.TweetID) > 0 {
			tw_client := twitter.NewClient(config)
			screenname, url := tw_client.GetScreennameAndLink(url.TweetID)
			if len(screenname) > 0 && len(url) > 0 {
				tweet_add = fmt.Sprintf("I scanned this repository after [@%s](%s) requested it on Twitter.", screenname, url)
			}
		}

		_, err = parser.PullRequest(branchname, "I fixed a few typos", tweet_add)
		if err != nil {
			log.Println("PullRequest failed:", err)
			return
		}
	}
}

func ParseGitHubURL(rawurl string) (*GitHubURL, error) {
	parsed, err := netUrl.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	if parsed.Host != "github.com" {
		return nil, errors.New("Not a GitHub URL")
	}

	splitted := strings.Split(parsed.Path, "/")
	owner := splitted[1]
	repository := splitted[2]

	return &GitHubURL{
		URL:        parsed,
		Owner:      owner,
		Repository: repository,
	}, nil
}
