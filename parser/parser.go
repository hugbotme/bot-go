package parser

import (
	"bufio"
	"errors"
	"github.com/google/go-github/github"
	"github.com/hugbotme/bot-go/config"
	"github.com/hugbotme/bot-go/repository"
	"github.com/libgit2/git2go"
	"log"
	"os"
	"strings"
	"time"
	"io/ioutil"
)

type Parser struct {
	client             *repository.GithubClient
	configuration      *config.Configuration
	clonedProjectsPath string
	username           string
	repositoryname     string
	signature          *git.Signature
	repopointer        *git.Repository
}

func (p *Parser) GetClonedProjectsPath() string {
	return p.clonedProjectsPath
}
func (p *Parser) GetRepositoryname() string {
	return p.repositoryname
}

func (p *Parser) Clone(repo *github.Repository) error {
	os.RemoveAll(p.clonedProjectsPath + p.repositoryname)
	repopointer, err := git.Clone(*repo.CloneURL, p.clonedProjectsPath+p.repositoryname, &git.CloneOptions{})
	p.repopointer = repopointer

	if err != nil {
		return err
	}
	return nil
}

// Init function to define arguments
func NewParser(username, repositoryname string, config *config.Configuration) Parser {

	client := repository.NewGithubClient(&config.Github)
	clonedProjectsPath := "cloned_projects/"

	signature := &git.Signature{
		Name:  config.Git.Name,
		Email: config.Git.Email,
		When:  time.Now(),
	}

	return Parser{
		client:             client,
		configuration:      config,
		clonedProjectsPath: clonedProjectsPath,
		username:           username,
		repositoryname:     repositoryname,
		signature:          signature,
		repopointer:        nil,
	}
}

func (p Parser) ForkRepository() (*github.Repository, *github.Response, error) {
	// list all repositories for the authenticated user
	//repos, _, err := githubClient.Client.Repositories.List("", nil)

	// Get contents of README of a repo
	// if err...

	forkOptions := github.RepositoryCreateForkOptions{
		"",
	}

	return p.client.Client.Repositories.CreateFork(p.username, p.repositoryname, &forkOptions)
}

func (p Parser) GetReadme() (string, []string, error) {

	readmeFiles := []string{"README.md", "README.txt", "README", "Readme.md", "Readme.txt", "Readme"}

	for _, filename := range readmeFiles {
		path := p.clonedProjectsPath + p.repositoryname + "/" + filename
		if _, err := os.Stat(path); err == nil {
			log.Printf("Readme file exists; processing...")
			lines, err := p.GetFileContents(filename)
			return filename, lines, err
		}
	}

	return "", nil, errors.New("Could not find README file :,(")
}

func (p Parser) GetFileContents(filename string) ([]string, error) {
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

	return p.repopointer.CreateBranch(branchname, headCommit, false, p.signature, "Branch for "+branchname)
}

func (p Parser) CommitFile(branch *git.Branch, branchname string, filename string, contents string, msg string) error {
	filepath := p.clonedProjectsPath + p.repositoryname + "/" + filename
	ioutil.WriteFile(filepath, []byte(contents), 0644)
	treeId, err := p.AddFilePath(filename)

	if err != nil {
		return err
	}
	return p.Commit(branch, branchname, treeId, msg)
}

func (p Parser) PullRequest(branchname, msg, tweet_add string) (*github.PullRequest, error) {

	cbs := &git.RemoteCallbacks{
		CredentialsCallback: func(url string, username string, allowedTypes git.CredType) (git.ErrorCode, *git.Cred) {
			ret, cred := git.NewCredUserpassPlaintext(p.configuration.Github.Username, p.configuration.Github.APIToken)
			return git.ErrorCode(ret), &cred
		},
		CertificateCheckCallback: func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
			if hostname != "github.com" {
				return git.ErrUser
			}
			return 0
		},
	}

	user := p.configuration.Github.Username
	remote := "https://github.com/" + user + "/" + p.repositoryname + ".git"
	log.Println("remote", remote)
	fork, err := p.repopointer.CreateRemote("fork", remote)
	if err != nil {
		return nil, err
	}

	err = fork.SetCallbacks(cbs)
	if err != nil {
		return nil, err
	}

	err = fork.Push([]string{"refs/heads/" + branchname}, nil, p.signature, msg)

	if err != nil {
		return nil, err
	}

	log.Println("Pushed to", branchname)

	head := user + ":" + branchname
	base := "master"

	title := p.configuration.Github.PRTemplate.Title
	title = strings.Replace(title, "%title%", msg, 1)

	body := strings.Join(p.configuration.Github.PRTemplate.Body, "\n")

	commitmsg := "I found some typos in your README. I fixed them, I hope this helps.\n"
	commitmsg += "\n"

	if len(tweet_add) > 0 {
		commitmsg += "\n"
		commitmsg += tweet_add
	}

	// TODO: Maybe we'd like to do some fancy templating?
	body = strings.Replace(body, "%commit-msg%", commitmsg, 1)
	//body = strings.Replace(body, "%url%", m.Change.URL, 1)

	pr := &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
	}

	log.Println("new PullRequest", *pr, head, base)

	// Do the pull request itself
	prResult, resp, err := p.client.Client.PullRequests.Create(p.username, p.repositoryname, pr)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return prResult, nil
}

func (p Parser) Commit(branch *git.Branch, branchname string, treeId *git.Oid, msg string) error {
	tree, err := p.repopointer.LookupTree(treeId)
	if err != nil {
		return err
	}

	commitTarget, err := p.repopointer.LookupCommit(branch.Target())
	if err != nil {
		return err
	}

	_, err = p.repopointer.CreateCommit("refs/heads/"+branchname, p.signature, p.signature, msg, tree, commitTarget)
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
