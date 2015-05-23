package parser

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"github.com/hugbotme/bot-go/repository"
	"github.com/hugbotme/bot-go/config"
	"github.com/google/go-github/github"
)

var (
	flagConfigFile *string
	flagPidFile    *string
	flagVersion    *bool
)

type Parser struct {
	client *repository.GithubClient
}

const (
	majorVersion = 1
	minorVersion = 0
	patchVersion = 0
)

// Init function to define arguments
func NewParser() Parser {
	flagConfigFile = flag.String("config", "", "Configuration file")
	flagPidFile = flag.String("pidfile", "", "Write the process id into a given file")
	flagVersion = flag.Bool("version", false, "Outputs the version number and exits")

	config, err := getConfig()

	if err != nil {
		log.Fatal("Configuration initialisation failed:", err)
	}

	log.Println(config.Github)

	return Parser {
		client : repository.NewGithubClient(&config.Github),
	}

}

func (p Parser) ForkRepository(username, repo string) {
	// list all repositories for the authenticated user
	//repos, _, err := githubClient.Client.Repositories.List("", nil)


	// Get contents of README of a repo
	// if err...

	forkOptions := github.RepositoryCreateForkOptions {
		"hugbotme",
	}

	p.client.Client.Repositories.CreateFork(username, repo, &forkOptions)

}






func getConfig() (*config.Configuration, error)  {

	flag.Parse()

	// Check for configuration file
	if len(*flagConfigFile) <= 0 {
		log.Println("No configuration file found. Please add the --config parameter")
		return nil, errors.New("No configuration file found")
	}

	// PID-File
	if len(*flagPidFile) > 0 {
		ioutil.WriteFile(*flagPidFile, []byte(strconv.Itoa(os.Getpid())), 0644)
	}

	// Bootstrap configuration file
	return config.NewConfiguration(flagConfigFile)
}
