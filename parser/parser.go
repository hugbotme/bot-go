package parser

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/hugbotme/bot-go/config"
	"github.com/hugbotme/bot-go/repository"
	"github.com/libgit2/git2go"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

var (
	flagConfigFile *string
	flagPidFile    *string
	flagVersion    *bool
)

type Parser struct {
	client             *repository.GithubClient
	clonedProjectsPath string
	username           string
	repositoryname     string
}

const (
	majorVersion = 1
	minorVersion = 0
	patchVersion = 0
)

// Init function to define arguments
func NewParser(username, repositoryname string) Parser {
	flagConfigFile = flag.String("config", "", "Configuration file")
	flagPidFile = flag.String("pidfile", "", "Write the process id into a given file")
	flagVersion = flag.Bool("version", false, "Outputs the version number and exits")

	config, err := getConfig()

	if err != nil {
		log.Fatal("Configuration initialisation failed:", err)
	}

	log.Println(config.Github)

	client := repository.NewGithubClient(&config.Github)
	clonedProjectsPath := "cloned_projects/"

	return Parser{
		client:             client,
		clonedProjectsPath: clonedProjectsPath,
		username:           username,
		repositoryname:     repositoryname,
	}
}

func (p Parser) ForkRepository(username, repo string) (*github.Repository, *github.Response, error) {
	// list all repositories for the authenticated user
	//repos, _, err := githubClient.Client.Repositories.List("", nil)

	// Get contents of README of a repo
	// if err...

	forkOptions := github.RepositoryCreateForkOptions{
		"",
	}

	return p.client.Client.Repositories.CreateFork(username, repo, &forkOptions)
}

func (p Parser) GetFileContents(filename string) ([]string, error) {
	repo, _, err := p.ForkRepository(p.username, p.repositoryname)

	if err != nil {
		log.Printf("Error during fork: %v\n", err)
	}

	log.Printf("Forked repo:" + *repo.CloneURL)

	repoClone, err := git.Clone(*repo.CloneURL, p.clonedProjectsPath+p.repositoryname, &git.CloneOptions{})

	if err != nil {
		log.Printf("Error during clone: %v\n", err)
	}

	log.Printf("%v", repoClone)

	return p.ReadLines(p.clonedProjectsPath + p.repositoryname + "/" + filename)
}

func (p Parser) CommitFile(filename string, contents []string) {
	fmt.Println("Commit " + filename)
}


// readLines reads a whole file into memory
// and returns a slice of its lines.
func (p Parser) ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func (p Parser) WriteLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func getConfig() (*config.Configuration, error) {

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
