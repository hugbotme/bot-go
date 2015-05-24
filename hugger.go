package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hugbotme/bot-go/config"
	"github.com/hugbotme/bot-go/parser"
	"log"
	netUrl "net/url"
	"strings"
)

type GitHubURL struct {
	URL        *netUrl.URL
	Owner      string
	Repository string
}

func processHug(url *Hug, config *config.Configuration) {
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
	lines, err := parser.GetReadme()

	var buffer bytes.Buffer
	if err != nil {
		log.Printf("Error reading README: %v\n", err)
	} else {
		for i, line := range lines {
			fmt.Println(i, line)
			buffer.WriteString(line)
		}

		processor, err := newSpellCheckFileProcessor()
		if err != nil {
			fmt.Errorf("could not get speller: %s", err.Error())
			return
		}

		content := processor.processContent([]byte(buffer.String()))

		branchname := "bugfix"

		// TODO: ERROR HANDLING
		branch, err := parser.CreateBranch(branchname)
		parser.CommitFile(branch, branchname, "Readme.md", content, "Fixing some typos")
		parser.PullRequest(branchname, "A friendly pull request")
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
