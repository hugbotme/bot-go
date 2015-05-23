package parser

import (
	"bufio"
	"errors"
	"time"
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
	configuration		*config.Configuration
	clonedProjectsPath string
	username           string
	repositoryname     string
	signature *git.Signature
	repopointer *git.Repository
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

	client := repository.NewGithubClient(&config.Github)
	clonedProjectsPath := "cloned_projects/"

	signature := &git.Signature {
		Name: config.Git.Name,
		Email: config.Git.Email,
		When: time.Now(),
	}

	return Parser{
		client:             client,
		configuration: 		config,
		clonedProjectsPath: clonedProjectsPath,
		username:           username,
		repositoryname:     repositoryname,
		signature: signature,
		repopointer: nil,
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

	repopointer, err := git.Clone(*repo.CloneURL, p.clonedProjectsPath+p.repositoryname, &git.CloneOptions{})
	p.repopointer = repopointer

	if err != nil {
		log.Printf("Error during clone: %v\n", err)
	}

	return p.ReadLines(p.clonedProjectsPath + p.repositoryname + "/" + filename)
}

func (p Parser) CreateBranch(branchname string) (*git.Branch, error) {

	head, err := p.repopointer.Head()
	if err != nil {
		return nil, err
	}

	headCommit, err := p.repopointer.LookupCommit(head.Target())
	if err != nil {
		return nil, err
	}

	return p.repopointer.CreateBranch(branchname, headCommit, false, p.signature, "Branch for " + branchname)
}

func (p Parser) CommitFile(branch *git.Branch, branchname string, filename string, contents []string, msg string) error {
	filepath := p.clonedProjectsPath + p.repositoryname + "/" + filename

	p.WriteLines(filepath, contents)
	treeId, err := p.AddFilePath(filepath)

	if err != nil {
		return err
	}
	return p.Commit(branch, branchname, treeId, msg)
}

func (p Parser) PullRequest(msg string) {

}

func (p Parser) Commit(branch *git.Branch, branchname string, treeId *git.Oid, msg string) (error) {
	tree, err := p.repopointer.LookupTree(treeId)
	if err != nil {
return err
	}

	commitTarget, err := p.repopointer.LookupCommit(branch.Target())
	if err != nil {
		return err
	}

	_, err = p.repopointer.CreateCommit("refs/heads/" + branchname, p.signature, p.signature, msg, tree, commitTarget)
	return err
}


func (p Parser) AddFilePath(filepath string) (*git.Oid, error) {
	idx, err := p.repopointer.Index()
if err != nil {
return nil, err
}

err = idx.AddByPath(filepath)
if err != nil {
return nil, err
}

treeId, err := idx.WriteTree()
if err != nil {
return nil, err
}

err = idx.Write()
if err != nil {
return nil, err
}
	return treeId, nil
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
func (p Parser) WriteLines(path string, lines []string) error {
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
